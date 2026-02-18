package main

import (
    "github.com/playwright-community/playwright-go"
)

func main() {
    err := playwright.Install()
    if err != nil {
        panic(err)
    }
}