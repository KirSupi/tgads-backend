package coingecko

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/timmbarton/utils/tracing"
	"github.com/timmbarton/utils/types/dates"
	"resty.dev/v3"
)

type Client struct {
	c *resty.Client
}

func New(apiKey string) *Client {
	c := resty.New()

	c.SetHeader("x-coingecko-api-key", apiKey)
	c.SetBaseURL("https://api.coingecko.com/api/v3/")

	return &Client{
		c: c,
	}
}

type getTonRateResponse struct {
	Id           string `json:"id"`
	Symbol       string `json:"symbol"`
	Name         string `json:"name"`
	Localization struct {
		En string `json:"en"`
		Ru string `json:"ru"`
	} `json:"localization"`
	Image struct {
		Thumb string `json:"thumb"`
		Small string `json:"small"`
	} `json:"image"`
	MarketData struct {
		CurrentPrice struct {
			Rub decimal.Decimal `json:"rub"`
			Usd decimal.Decimal `json:"usd"`
		} `json:"current_price"`
		MarketCap struct {
			Rub decimal.Decimal `json:"rub"`
			Usd decimal.Decimal `json:"usd"`
		} `json:"total_volume"`
	} `json:"market_data"`
	CommunityData struct {
		FacebookLikes            decimal.Decimal `json:"facebook_likes"`
		TwitterFollowers         decimal.Decimal `json:"twitter_followers"`
		RedditAveragePosts48H    decimal.Decimal `json:"reddit_average_posts_48h"`
		RedditAverageComments48H decimal.Decimal `json:"reddit_average_comments_48h"`
		RedditSubscribers        decimal.Decimal `json:"reddit_subscribers"`
		RedditAccountsActive48H  decimal.Decimal `json:"reddit_accounts_active_48h"`
	} `json:"community_data"`
	DeveloperData struct {
		Forks                        int `json:"forks"`
		Stars                        int `json:"stars"`
		Subscribers                  int `json:"subscribers"`
		TotalIssues                  int `json:"total_issues"`
		ClosedIssues                 int `json:"closed_issues"`
		PullRequestsMerged           int `json:"pull_requests_merged"`
		PullRequestContributors      int `json:"pull_request_contributors"`
		CodeAdditionsDeletions4Weeks struct {
			Additions int `json:"additions"`
			Deletions int `json:"deletions"`
		} `json:"code_additions_deletions_4_weeks"`
		CommitCount4Weeks int `json:"commit_count_4_weeks"`
	} `json:"developer_data"`
	PublicInterestStats struct {
		AlexaRank   any `json:"alexa_rank"`
		BingMatches any `json:"bing_matches"`
	} `json:"public_interest_stats"`
}

func (c *Client) GetTonRate(ctx context.Context, date dates.Date) (rate decimal.Decimal, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer span.End()

	resp, err := c.c.R().
		SetContext(ctx).
		SetQueryParam("date", time.Time(date).Format("02-01-2006")).
		SetPathParam("id", "the-open-network").
		Get("/coins/{id}/history")
	if err != nil {
		return rate, err
	}

	if resp.StatusCode() != http.StatusOK {
		fmt.Println(resp.StatusCode(), resp.Status())
		fmt.Println(string(resp.Bytes()))
		return rate, errors.New("status code is not 200")
	}

	data := getTonRateResponse{}

	err = json.Unmarshal(resp.Bytes(), &data)
	if err != nil {
		return rate, err
	}

	rate = data.MarketData.CurrentPrice.Usd

	return rate, nil
}
