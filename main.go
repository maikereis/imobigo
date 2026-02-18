package main

import (
	"math/rand"
	"time"
	"log"
	"fmt"
	"github.com/playwright-community/playwright-go"
)

const URL = "https://www.vivareal.com.br/aluguel/rio-grande-do-sul/porto-alegre/casa_residencial/"

func RandomChoice[T any](r *rand.Rand, items []T) T{
	return items[r.Intn(len(items))]
}

func generateContext(browser playwright.Browser) (playwright.BrowserContext, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/144.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:147.0) Gecko/20100101 Firefox/147.0",
		"Mozilla/5.0 (X11; Linux i686; rv:147.0) Gecko/20100101 Firefox/147.0",
	}

	sizes := []playwright.Size{
		{Width: 1280, Height: 720},
		{Width: 1366, Height: 768},
		{Width: 1920, Height: 1080},
		{Width: 1440, Height: 900},
	}

	timezoneIds := []string{
		"America/Belem",
		"America/Sao_Paulo",
	}

	randomUserAgent := userAgents[r.Intn(len(userAgents))]
	randomSize := sizes[r.Intn(len(sizes))]
	randomTimezoneId := timezoneIds[r.Intn(len(timezoneIds))]

	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent:  playwright.String(randomUserAgent),
		Viewport:   &randomSize,
		Locale:     playwright.String("pt-BR"),
		TimezoneId: playwright.String(randomTimezoneId),
	})

	if err != nil {
		return nil, err
	}

	return context, nil
}


func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		//Headless: playwright.Bool(false),
		Channel:  playwright.String("chrome"),
		SlowMo:   playwright.Float(200),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	context, err := generateContext(browser)

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto(URL); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	ulLocator := page.Locator("xpath=/html/body/section/div[1]/div[3]/div[4]/div[1]/ul")

	err = ulLocator.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		log.Fatalf("advertisements list not found: %v", err)
	}

	liLocators, err := ulLocator.Locator("li").All()
	if err != nil {
		log.Fatalf("err trying to fetch cards: %v", err)
	}

	fmt.Printf("found %d advertisements\n", len(liLocators))

	for i, card := range liLocators {
		expandBtn := card.Locator("button:has-text('Ver'), a:has-text('Ver')")
		btnCount, _ := expandBtn.Count()

		if btnCount > 0 {
			fmt.Printf("Has expand button with %d advertisements in card %d\n", btnCount, i)
		} else {
			fmt.Printf("No expand button found in card %d\n", i)
		}
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
