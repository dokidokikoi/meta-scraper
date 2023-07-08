package scraper

type FanzaGames struct {
	Proxy     string
	Domain    string
	SearchUri string
}

var FanzaGamesScraper *FanzaGames

func init() {
	FanzaGamesScraper = &FanzaGames{
		Proxy:     "socks5://127.0.0.1:7890",
		Domain:    "https://dlsoft.dmm.co.jp/",
		SearchUri: "https://dlsoft.dmm.co.jp/search?service=pcgame&floor=digital_pcgame&searchstr=%d",
	}
}
