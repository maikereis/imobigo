package main

import (
	"log"
	"fmt"
	"github.com/playwright-community/playwright-go"
)

const URL = "https://www.vivareal.com.br/aluguel/rio-grande-do-sul/porto-alegre/casa_residencial/"

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		//Headless: playwright.Bool(false),
		Channel:  playwright.String("chrome"),
		SlowMo:   playwright.Float(500),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		Viewport: &playwright.Size{Width: 1280, Height: 720},
		Locale:   playwright.String("pt-BR"),
    	TimezoneId: playwright.String("America/Belem"),
	})


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
		text, err := card.InnerText()
		if err != nil {
			log.Printf("err trying to read card %d: %v", i, err)
			continue
		}
		fmt.Printf("card %d: %s\n", i+1, text)
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}
