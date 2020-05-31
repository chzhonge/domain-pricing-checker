package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

func getConfig(disableGui bool) []chromedp.ExecAllocatorOption {
	options := []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", disableGui),
		chromedp.Flag("hide-scrollbars", false),
		chromedp.Flag("mute-audio", false),
		chromedp.UserAgent(`Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36`),
	}

	return options
}

func runJob(url string, patterns []string) []string {
	ctx := context.Background()
	options := append(chromedp.DefaultExecAllocatorOptions[:], getConfig(true)...)
	c, cc := chromedp.NewExecAllocator(ctx, options...)
	defer cc()

	ctx, cancel := chromedp.NewContext(c)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()

	datas := make([]string, len(patterns))

	tasks := chromedp.Tasks{
		chromedp.Navigate(url),
	}

	for index, pattern := range patterns {
		if pattern == ".tooltip-content" {
			tasks = append(tasks, chromedp.Click(".uxicon.uxicon-help"))
		}
		tasks = append(tasks, chromedp.WaitVisible(pattern, chromedp.ByQuery))
		tasks = append(tasks, chromedp.TextContent(pattern, &datas[index], chromedp.ByQuery))
	}

	chromedp.Run(ctx, tasks)

	return datas
}

func getGandiPricing(domainName string) []string {
	patterns := []string{
		`.Text-align-end_3rH75.PriceText-topPrice__2UzVP`,
		`.Text-size-small_wppb0.Text-align-end_3rH75.PriceText-bottomPrice__3uubF`,
	}

	result := runJob(
		"https://shop.gandi.net/zh-hant/domain/suggest?search="+domainName,
		patterns,
	)

	for i, r := range result {
		tmp := strings.ReplaceAll(r, " ", "")
		match := regexp.MustCompile(`(?m)NT\$\S\d*`).FindString(tmp)
		if match != "" {
			result[i] = match
		}
	}

	return result
}

func getGodaddyPricing(domainName string) []string {
	patterns := []string{
		`.h3.text-primary.dpp-price.m-b-0.ds-dpp-price.ds-intl`,
		`.tooltip-content`,
	}

	result := runJob(
		"https://tw.godaddy.com/domainsearch/find?checkAvail=1&tmskey=&domainToCheck="+domainName,
		patterns,
	)

	match := regexp.MustCompile(`(?m)NT\$\d*`).FindString(result[1])
	if match != "" {
		result[1] = match
	}

	return result
}

func main() {
	domainName := "gomanners.com"

	gandiPricing := getGandiPricing(domainName)
	godaddyPricing := getGodaddyPricing(domainName)

	fmt.Printf("Gandi 價格 首年:%s 歷年:%s \n", gandiPricing[0], gandiPricing[1])
	fmt.Printf("Godaddy 價格 首年:%s 歷年:%s \n", godaddyPricing[0], godaddyPricing[1])
}
