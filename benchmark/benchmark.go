package benchmark

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mittz/roleplay-webapp-assess/product"
	"golang.org/x/sync/errgroup"
)

const (
	BENCHMARK_TIMEOUT_SECOND          = 60
	NUM_OF_BENCHMARKER                = 4
	MAX_PRODUCT_QUANTITY_PER_CHECKOUT = 100
	SCORE_GET_PRODUCTS                = 5
	SCORE_POST_CHECKOUT               = 2
	SCORE_GET_PRODUCT                 = 1
	SCORE_GET_CHECKOUTS               = 4
)

var httpClient *http.Client

func Run(userkey, endpoint string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*BENCHMARK_TIMEOUT_SECOND)
	eg, ctx := errgroup.WithContext(ctx)
	defer cancel()

	scores := make(chan int)
	for i := 0; i < NUM_OF_BENCHMARKER; i++ {
		eg.Go(func() error {
			return benchmark(ctx, endpoint, scores)
		})
	}

	totalScore := 0
	for i := 0; i < NUM_OF_BENCHMARKER; i++ {
		totalScore += <-scores
	}

	if err := eg.Wait(); err != nil {
		return totalScore, err
	}

	return totalScore, nil
}

func benchmark(ctx context.Context, endpoint string, score chan<- int) error {
	total := 0
	for {
		select {
		case <-ctx.Done():
			score <- total
			return nil
		default: // do benchmark
			rand.Seed(time.Now().UnixNano())
			baseURL, err := url.Parse(endpoint)
			if err != nil {
				log.Printf("%v\n", err)
			}

			productID := rand.Intn(product.GetNumOfProducts()-1) + 1              // Exclude 0
			productQuantity := rand.Intn(MAX_PRODUCT_QUANTITY_PER_CHECKOUT-1) + 1 // Exclude 0

			result := benchGetProducts(*baseURL)
			if result == 0 {
				score <- 0
				return fmt.Errorf("unable to get an expected result from GET /products")
			}
			total += result

			result = benchPostCheckout(*baseURL, productID, productQuantity)
			if result == 0 {
				score <- 0
				return fmt.Errorf("unable to get an expected result from POST /checkout")
			}
			total += result

			result = benchGetProduct(*baseURL)
			if result == 0 {
				score <- 0
				return fmt.Errorf("unable to get an expected result from GET /product")
			}
			total += result

			result = benchGetCheckouts(*baseURL, productID, productQuantity)
			if result == 0 {
				score <- 0
				return fmt.Errorf("unable to get an expected result from GET /checkouts")
			}
			total += result
		}
	}
}

func newHTTPClient() *http.Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				MaxConnsPerHost: 20,
			},
		}
	}

	return httpClient
}

func benchGetProducts(baseURL url.URL) int {
	getProductsURL := baseURL
	getProductsURL.Path = path.Join(getProductsURL.Path, "/products")
	httpClient := newHTTPClient()
	resp, err := httpClient.Get(getProductsURL.String())
	if err != nil {
		log.Printf("%v\n", err)
		return 0
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("%v\n", err)
		return 0
	}

	// Check hashsum of image
	imagePaths := doc.Find("div.content-container").Find("img.card-img-top.products-img").Map(func(_ int, s *goquery.Selection) string {
		val, _ := s.Attr("src")

		if !strings.HasPrefix(val, "http") {
			return fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, val)
		}

		return val
	})

	productID := rand.Intn(product.GetNumOfProducts()-1) + 1 // Exclude 0
	if len(imagePaths) <= productID {
		return 0
	}

	imagePath := imagePaths[productID]
	respImage, err := http.Get(imagePath)
	if err != nil {
		log.Printf("%v\n", err)
		return 0
	}
	defer respImage.Body.Close()

	h := md5.New()
	if _, err := io.Copy(h, respImage.Body); err != nil {
		log.Printf("%v\n", err)
		return 0
	}

	// TODO: Check stylesheets
	if resp.StatusCode == http.StatusOK && fmt.Sprintf("%x", h.Sum(nil)) == product.GetImageHash(path.Base(imagePath)) {
		return SCORE_GET_PRODUCTS
	}

	return 0
}

func benchPostCheckout(baseURL url.URL, productID int, productQuantity int) int {
	data := url.Values{
		"product_id":       {fmt.Sprintf("%d", productID)},
		"product_quantity": {fmt.Sprintf("%d", productQuantity)},
	}
	postCheckout := baseURL
	postCheckout.Path = path.Join(postCheckout.Path, "/checkout")
	httpClient := newHTTPClient()
	resp, err := httpClient.PostForm(postCheckout.String(), data)
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0
	}

	orderInfo := doc.Find("div.content-container").Find("p.card-text").Text()
	imagePath, ok := doc.Find("div.content-container").Find("img.checkout-img").Attr("src")
	if !ok {
		return 0
	}

	if !strings.HasPrefix(imagePath, "http") {
		imagePath = fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, imagePath)
	}

	respImage, err := httpClient.Get(imagePath)
	if err != nil {
		log.Printf("%v\n", err)
		return 0
	}
	defer respImage.Body.Close()

	h := md5.New()
	if _, err := io.Copy(h, respImage.Body); err != nil {
		log.Printf("%v\n", err)
		return 0
	}

	if resp.StatusCode == http.StatusAccepted &&
		strings.Contains(orderInfo, fmt.Sprintf("%d x", productQuantity)) &&
		fmt.Sprintf("%x", h.Sum(nil)) == product.GetImageHash(path.Base(imagePath)) {
		return SCORE_POST_CHECKOUT
	}

	return 0
}

func benchGetProduct(baseURL url.URL) int {
	getProductURL := baseURL
	productID := rand.Intn(product.GetNumOfProducts()-1) + 1 // Exclude 0
	getProductURL.Path = path.Join(getProductURL.Path, "/product", fmt.Sprintf("%d", productID))
	httpClient := newHTTPClient()
	resp, err := httpClient.Get(getProductURL.String())
	if err != nil {
		log.Printf("%v\n", err)
		return 0
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("%v\n", err)
		return 0
	}

	imagePath, ok := doc.Find("div.content-container").Find("img.product-img").Attr("src")
	if !ok {
		return 0
	}

	if !strings.HasPrefix(imagePath, "http") {
		imagePath = fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, imagePath)
	}

	respImage, err := httpClient.Get(imagePath)
	if err != nil {
		log.Printf("%v\n", err)
		return 0
	}
	defer respImage.Body.Close()

	h := md5.New()
	if _, err := io.Copy(h, respImage.Body); err != nil {
		log.Printf("%v\n", err)
		return 0
	}

	if resp.StatusCode == http.StatusOK && fmt.Sprintf("%x", h.Sum(nil)) == product.GetImageHash(path.Base(imagePath)) {
		return SCORE_GET_PRODUCT
	}

	return 0
}

func benchGetCheckouts(baseURL url.URL, productID int, productQuantity int) int {
	getCheckoutsURL := baseURL
	getCheckoutsURL.Path = path.Join(getCheckoutsURL.Path, "/checkouts")
	httpClient := newHTTPClient()
	resp, err := httpClient.Get(getCheckoutsURL.String())
	if err != nil {
		return 0
	}
	defer resp.Body.Close()

	// Check if the order which is just created exists
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return 0
	}

	order := doc.Find("table").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		return s.Find("td.product_id").Text() != fmt.Sprint(productID) ||
			s.Find("td.product_quantity").Text() != fmt.Sprint(productQuantity)
	})

	imagePath, ok := order.Find("td.product_image").Find("img").Attr("src")
	if !ok {
		return 0
	}

	if !strings.HasPrefix(imagePath, "http") {
		imagePath = fmt.Sprintf("%s://%s%s", baseURL.Scheme, baseURL.Host, imagePath)
	}

	respImage, err := httpClient.Get(imagePath)
	if err != nil {
		log.Printf("%v\n", err)
		return 0
	}
	defer respImage.Body.Close()

	h := md5.New()
	if _, err := io.Copy(h, respImage.Body); err != nil {
		log.Printf("%v\n", err)
		return 0
	}

	if resp.StatusCode == http.StatusOK && order != nil && fmt.Sprintf("%x", h.Sum(nil)) == product.GetImageHash(path.Base(imagePath)) {
		return SCORE_GET_CHECKOUTS
	}

	return 0
}
