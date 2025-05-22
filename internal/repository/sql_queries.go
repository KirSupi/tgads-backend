package repository

const (
	queryCreateStats = `
		INSERT INTO tgads.stats(campaign_id, "date", views, clicks, actions, spend, cpm)
		VALUES ($1::text,
				unnest($2::date[]),
				unnest($3::int[]),
				unnest($4::int[]),
				unnest($5::int[]),
				unnest($6::decimal[]),
				unnest($7::decimal[]))
		ON CONFLICT (campaign_id, "date") 
		             DO UPDATE SET 
		                 views = EXCLUDED.views, 
		                 clicks = EXCLUDED.clicks, 
		                 actions = EXCLUDED.actions, 
		                 spend = EXCLUDED.spend, 
		                 cpm = EXCLUDED.cpm
	`
	queryCreateCampaign = `
		INSERT INTO tgads.campaigns(id, name, stats_csv_link, budget_csv_link, text, button_text, link, active)
		VALUES ($1::text, 
		        $2::text,
				$3::text,
				$4::text,
				$5::text,
				$6::text,
				$7::text,
				$8::boolean)
		ON CONFLICT (id) DO NOTHING
	`
	queryFetchCampaigns = `
		SELECT *
		FROM tgads.campaigns
	`
	queryCreateRate = `
		INSERT INTO tgads.rates(date, rate)
		VALUES ($1::date,$2::decimal)
		ON CONFLICT (DATE) DO UPDATE SET rate = EXCLUDED.rate
	`
)
