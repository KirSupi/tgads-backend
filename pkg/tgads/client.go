package tgads

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/icrowley/fake"
	"github.com/shopspring/decimal"
	"github.com/timmbarton/utils/tracing"
	"resty.dev/v3"
)

type Client struct {
	c *resty.Client
}

func New() *Client {
	c := resty.New()
	c.SetHeader("User-Agent", fake.UserAgent())

	return &Client{
		c: c,
	}
}

var (
	linksRegex      = regexp.MustCompile(`"csvExport":"\\`)
	campaignIdRegex = regexp.MustCompile(`^https://ads\.telegram\.org/stats/([A-Za-z0-9]+)$`)
)

func GetCampaignShareLink(id string) (link string) {
	return fmt.Sprintf("https://ads.telegram.org/stats/%s", id)
}

func GetCampaignId(link string) (id string, err error) {
	matches := campaignIdRegex.FindStringSubmatch(link)
	if matches == nil {
		return id, errors.New("invalid link")
	}

	id = matches[1]

	return id, nil
}

type Campaign struct {
	Id            string `json:"id"`
	StatsCSVLink  string `json:"stats_csv_link"`
	BudgetCSVLink string `json:"budget_csv_link"`
	Text          string `json:"text"`
	ButtonText    string `json:"button_text"`
	Link          string `json:"link"`
	Active        bool   `json:"active"`
}

func (c *Client) GetCampaign(ctx context.Context, link string) (res Campaign, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	resp, err := c.c.R().SetContext(ctx).Get(link)
	if err != nil {
		return res, err
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return res, err
	}

	// Проверяем на валидность
	campaignNotFound := doc.Find("meta[property=\"og:title\"]").Size() == 1
	if campaignNotFound {
		return res, errors.New("campaign not found")
	}

	// Id
	res.Id, err = GetCampaignId(link)
	if err != nil {
		return res, err
	}

	ok := false

	// Link
	res.Link, ok = doc.Find("div.pr-ad-info-value>a").Attr("href")
	if !ok {
		return res, errors.New("cant get link")
	}

	// Active
	status := strings.TrimSpace(doc.Find("div.pr-review-ad-info-multi").First().Find("div.pr-ad-info-value").First().Text())
	res.Active = status == "Active"

	// Text
	res.Text, err = doc.Find("div.ad-msg-link-preview-desc").Html()
	if err != nil {
		return res, err
	}

	// ButtonText
	res.ButtonText = doc.Find("div.ad-msg-link-preview-btn").Text()

	// StatsCSVLink
	body := resp.Bytes()

	indexes := linksRegex.FindAllIndex(body, -1)
	if len(indexes) != 2 {
		return res, errors.New("indexes count != 2")
	}

	startIndex := indexes[0][1]
	endIndex := bytes.Index(body[startIndex:], []byte("\""))
	res.StatsCSVLink = "https://ads.telegram.org" + string(body[startIndex:startIndex+endIndex])

	statsCSVLink, err := url.Parse(res.StatsCSVLink)
	if err != nil {
		return res, err
	}

	statsCSVLink.Scheme = "https"
	params := statsCSVLink.Query()
	params.Set("period", "day")
	statsCSVLink.RawQuery = params.Encode()
	res.StatsCSVLink = statsCSVLink.String()

	// BudgetCSVLink
	startIndex = indexes[1][1]
	endIndex = bytes.Index(body[startIndex:], []byte("\""))
	res.BudgetCSVLink = "https://ads.telegram.org" + string(body[startIndex:startIndex+endIndex])

	budgetCSVLink, err := url.Parse(res.BudgetCSVLink)
	if err != nil {
		return res, err
	}

	budgetCSVLink.Scheme = "https"
	params = budgetCSVLink.Query()
	params.Set("period", "day")
	budgetCSVLink.RawQuery = params.Encode()
	res.BudgetCSVLink = budgetCSVLink.String()

	return res, nil
}

type Stats struct {
	Datetime time.Time
	Views    int
	Clicks   int
	Actions  int
	Spend    decimal.Decimal
	CPM      decimal.Decimal
}

const (
	telegramDatetimeFormat = "02 Jan 2006" // "02 Jan 2006 15:04 MST"
	contentTypeCsv         = "text/csv"
	headerContentType      = "Content-Type"
)

var thousand = decimal.NewFromInt(1000)

func (c *Client) GetStats(ctx context.Context, statsLink, budgetLink string) (res []*Stats, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	rows, err := c.getTsv(ctx, statsLink)
	if err != nil {
		return res, err
	}

	res = make([]*Stats, 0)
	item := (*Stats)(nil)
	row := []string(nil)

	if len(rows) > 1 {
		if len(rows[0]) > 4 {
			return res, errors.New("stats table cols != 4")
		}

		for _, row = range rows[1:] {
			item = new(Stats)

			item.Datetime, err = time.Parse(telegramDatetimeFormat, row[0])
			if err != nil {
				return res, err
			}

			item.Views, err = strconv.Atoi(onlyNumeric(row[1]))
			if err != nil {
				return res, err
			}

			item.Clicks, err = strconv.Atoi(onlyNumeric(row[2]))
			if err != nil {
				return res, err
			}

			item.Actions, err = strconv.Atoi(onlyNumeric(row[3]))
			if err != nil {
				return res, err
			}

			res = append(res, item)
		}
	}

	rows, err = c.getTsv(ctx, budgetLink)
	if err != nil {
		return res, err
	}

	datetime := time.Time{}
	i := 0

	if len(rows) > 1 {
		if len(rows[0]) > 3 {
			return res, errors.New("budget table cols != 3")
		}

		for i, row = range rows[1:] {
			datetime, err = time.Parse(telegramDatetimeFormat, row[0])
			if err != nil {
				return res, err
			}

			item = res[i]

			if !item.Datetime.Equal(datetime) {
				return res, errors.New("invalid datetime")
			}

			item.Spend, err = decimal.NewFromString(strings.ReplaceAll(row[1], ",", "."))
			if err != nil {
				return res, err
			}

			if item.Views > 0 {
				item.CPM = item.Spend.Mul(thousand).Div(decimal.NewFromInt(int64(item.Views)))
			} else {
				item.CPM = item.Spend.Mul(thousand).Div(decimal.NewFromInt(1))
			}
		}
	}

	return res, nil
}

func (c *Client) getTsv(ctx context.Context, link string) (res [][]string, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	resp, err := c.c.R().SetContext(ctx).Get(link)
	if err != nil {
		return res, err
	}

	if resp.StatusCode() != http.StatusOK {
		return res, errors.New("status code is not 200")
	}

	if resp.Header().Get(headerContentType) != contentTypeCsv {
		return res, errors.New("content type is not csv")
	}

	r := csv.NewReader(bytes.NewBuffer(resp.Bytes()))
	r.Comma = '\t'

	res, err = r.ReadAll()
	if err != nil {
		return res, err
	}

	return res, nil
}

func onlyNumeric(in string) (out string) {
	for _, v := range in {
		if v >= '0' && v <= '9' {
			out += string(v)
		}
	}

	return out
}
