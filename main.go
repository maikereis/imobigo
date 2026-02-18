package main

import (
	"math/rand"
	"time"
	"log"
	"fmt"
	"github.com/playwright-community/playwright-go"
)

const URL = "https://www.vivareal.com.br/aluguel/rio-grande-do-sul/porto-alegre/casa_residencial/"

func RandomChoice[T any](r *rand.Rand, items []T) T {
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

func extractLinks(page playwright.Page) ([]string, error) {

	ulLocator := page.Locator("xpath=/html/body/section/div[1]/div[3]/div[4]/div[1]/ul")

	err := ulLocator.WaitFor(playwright.LocatorWaitForOptions{
		State: playwright.WaitForSelectorStateVisible,
	})

	if err != nil {
		log.Fatalf("advertisements list not found: %v", err)
	}

	liLocators, err := ulLocator.Locator(":scope > li").All()
	if err != nil {
		log.Fatalf("err trying to fetch cards: %v", err)
	}

	fmt.Printf("found %d advertisements\n", len(liLocators))

	var hrefs []string

	for i, card := range liLocators {
		// Check if this is an expandable (deduplicated) card by looking for the
		// "Ver os X anúncios deste imóvel" button.
		// Count() is synchronous and never times out — safe to use for existence checks.
		dedupButton := card.Locator(`button[data-cy="listing-card-deduplicated-button"]`)
		dedupCount, err := dedupButton.Count()
		if err != nil {
			log.Printf("card %d: error checking for dedup button: %v", i, err)
			continue
		}

		if dedupCount > 0 {
			// Expandable card: click the button to open the dialog
			fmt.Printf("Card %d: expandable card, opening dialog...\n", i)
			if err := dedupButton.Click(); err != nil {
				log.Printf("card %d: could not click dedup button: %v", i, err)
				continue
			}

			// Wait for the dialog to appear
			dialog := page.Locator(`div[role="dialog"][data-state="open"]`)
			if err := dialog.WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateVisible,
			}); err != nil {
				log.Printf("card %d: dialog did not appear: %v", i, err)
				continue
			}

			// Scrape all sub-card hrefs from the dialog.
			// Each sub-card is an <a href> inside the deduplication listings section.
			subCards, err := dialog.Locator(`section[data-cy="deduplication-modal-list-step"] a[href]`).All()
			if err != nil {
				log.Printf("card %d: could not get sub-cards: %v", i, err)
			} else {
				for j, subCard := range subCards {
					href, err := subCard.GetAttribute("href")
					if err != nil || href == "" {
						log.Printf("card %d sub-card %d: could not get href: %v", i, j, err)
						continue
					}
					fmt.Printf("Card %d sub-card %d href: %s\n", i, j, href)
					hrefs = append(hrefs, href)
				}
			}

			// Close the dialog by clicking the close button
			closeButton := dialog.Locator(`button[aria-label="Fechar modal"]`)
			if err := closeButton.Click(); err != nil {
				log.Printf("card %d: could not close dialog: %v", i, err)
				continue
			}

			// Wait for the dialog to close before moving on
			if err := dialog.WaitFor(playwright.LocatorWaitForOptions{
				State: playwright.WaitForSelectorStateHidden,
			}); err != nil {
				log.Printf("card %d: dialog did not close: %v", i, err)
			}
		} else {
			// Normal card: get href from the <a> tag.
			// Use Count() first to avoid a 30s timeout when the element is absent.
			aLocator := card.Locator(`a[href]`)
			aCount, err := aLocator.Count()
			if err != nil {
				log.Printf("card %d: error counting anchors: %v", i, err)
				continue
			}
			if aCount == 0 {
				log.Printf("card %d: no <a href> found, skipping", i)
				continue
			}
			href, err := aLocator.First().GetAttribute("href")
			if err != nil || href == "" {
				log.Printf("card %d: could not get href: %v", i, err)
				continue
			}
			fmt.Printf("Card %d href: %s\n", i, href)
			hrefs = append(hrefs, href)
		}
	}

	return hrefs, err
}

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("could not start playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Channel:  playwright.String("chrome"),
		SlowMo:   playwright.Float(200),
	})
	if err != nil {
		log.Fatalf("could not launch browser: %v", err)
	}

	context, err := generateContext(browser)
	if err != nil {
		log.Fatalf("could not generate context: %v", err)
	}

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = page.Goto(URL); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	//hrefs, err := extractLinks(page)
	hrefs := []string{"https://www.vivareal.com.br/imovel/casa-3-quartos-belem-velho-porto-alegre-com-garagem-12m2-aluguel-RS1750-id-2870890775/?source=ranking%2Crp"}

	fmt.Printf("\nTotal hrefs collected: %d\n", len(hrefs))
	fmt.Println(hrefs[0])

	pageToScrap, err := context.NewPage()
	if err != nil {
		log.Fatalf("could not create page: %v", err)
	}
	if _, err = pageToScrap.Goto(hrefs[0]); err != nil {
		log.Fatalf("could not goto: %v", err)
	}

	time.Sleep(2 * time.Second)

	title, err := pageToScrap.Locator("h2.text-neutral-130.font-semibold").First().InnerText()
	if err != nil {
		log.Printf("could not get title: %v", err)
	}
	fmt.Println("Title:", title)

	pageToScrap.Evaluate(`window.scrollTo(0, document.body.scrollHeight / 5)`)
	time.Sleep(1 * time.Second)

	_, err = pageToScrap.WaitForSelector(`[data-testid="amenities-list"]`)
	if err != nil {
		log.Fatalf("amenities list not found: %v", err)
	}

	amenitySpans, err := pageToScrap.Locator(`span.amenities-item-text`).All()
	if err != nil {
		log.Fatalf("could not get amenity spans: %v", err)
	}

	fmt.Printf("Found %d amenities\n", len(amenitySpans))
	for _, span := range amenitySpans {
		text, _ := span.InnerText()
		fmt.Println("Amenity:", text)
	}

	if err = browser.Close(); err != nil {
		log.Fatalf("could not close browser: %v", err)
	}
	if err = pw.Stop(); err != nil {
		log.Fatalf("could not stop Playwright: %v", err)
	}
}

