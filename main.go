package main

import (
	"bufio"
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/suzujun/youtube-puller/pkg/backoff"
)

var (
	bom = []byte{0xEF, 0xBB, 0xBF}
	JST = time.FixedZone("JST", 9*60*60)

	output     string
	timeout    time.Duration
	maxRetries int
)

const (
	defaultTimeout    = time.Second * 10
	defaultMaxRetries = 5
)

func main() {
	ctx := context.Background()
	defaultOutput := time.Now().Format("youtube_20060102.csv")
	flag.StringVar(&output, "output", defaultOutput, "output file path")
	flag.DurationVar(&timeout, "timeout", defaultTimeout, "request timeout")
	flag.IntVar(&maxRetries, "max-retry", defaultMaxRetries, "max retry count")
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		return
	}
	fmt.Println("----------")
	fmt.Println("YouTubeチャンネル情報の取得処理を開始します")
	for _, arg := range args {
		fmt.Println("読み込み:", arg)
	}
	fmt.Println("----------")
	err := run(ctx, args, output)
	fmt.Println("----------")
	if err != nil {
		fmt.Println("エラーが発生したため処理を中断します")
		fmt.Println("error:", err.Error())
	} else {
		fmt.Println("処理が正常に完了しました")
	}
	fmt.Println("何かキーを入力すると終了します")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
}

func run(ctx context.Context, args []string, output string) error {
	videos, err := listVideos(args)
	if err != nil {
		return err
	}
	tmpName := time.Now().Format("_tmp_youtube_20060102150405")
	file, err := os.Create(tmpName)
	if err != nil {
		return err
	}
	defer func() {
		file.Close()
		os.Remove(file.Name())
	}()
	if _, err := file.Write(bom); err != nil {
		return err
	}
	writer := csv.NewWriter(file)
	err = writer.Write([]string{"VideoURL", "ChannelURL", "ChannelTitle", "Error"})
	if err != nil {
		return err
	}
	defer writer.Flush()
	for i, video := range videos {
		fmt.Printf("%d,", i+1)
		title, channelURL, errmsg := getVideoInfomation(ctx, video)
		err = writer.Write([]string{video.path, channelURL, title, errmsg})
		if err != nil {
			return err
		}
	}
	fmt.Println("完了")
	writer.Flush()
	file.Close()
	return os.Rename(tmpName, output)
}

type Video struct {
	path   string
	errmsg string
	valid  bool
}

func listVideos(paths []string) ([]*Video, error) {
	videos := make([]*Video, 0, len(paths))
	for _, path := range paths {
		if strings.HasPrefix(path, "youtube.com/") {
			path = "https://www." + path
		}
		if !strings.HasPrefix(path, "https://") {
			res, err := readFile(path)
			if err != nil {
				return nil, err
			}
			videos = append(videos, res...)
			continue
		}
		video := &Video{path: path}
		u, err := url.Parse(path)
		switch {
		case err != nil:
			video.errmsg = err.Error()
		case u.Host != "www.youtube.com":
			video.errmsg = "Not YouTube URL"
		default:
			video.valid = true
		}
		videos = append(videos, video)
	}
	return videos, nil
}

func readFile(path string) ([]*Video, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	var paths []string
	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		paths = append(paths, scanner.Text())
	}
	urls, _ := listVideos(paths)
	return urls, scanner.Err()
}

var client = &http.Client{}

func getVideoInfomation(ctx context.Context, video *Video) (title, channelURL, errmsg string) {
	if !video.valid {
		errmsg = video.errmsg
		return
	}
	body, err := do(ctx, video.path)
	if err != nil {
		errmsg = err.Error()
		return
	}
	title, channelURL = parseInfomation(body)
	return
}

func do(ctx context.Context, url string) (body []byte, err error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var resp *http.Response
	backoff := backoff.NewExponentialBackoff(maxRetries)
	for backoff.Continue() {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		resp, err = client.Do(req.WithContext(ctx))
		switch {
		case errors.Is(err, context.DeadlineExceeded),
			errors.Is(err, context.Canceled):
			<-backoff.Wait()
			continue
		case err != nil:
			return
		}
		defer resp.Body.Close()
		body, err = io.ReadAll(resp.Body)
		return
	}
	err = fmt.Errorf("main: retry limit exceeded. %w", err)
	return
}

var (
	regexChannelTitle = regexp.MustCompile(`<link itemprop="name" content="([^"]+)">`)
	regexChannelURL   = regexp.MustCompile(`"canonicalBaseUrl":"/(channel/|@)([a-zA-Z0-9_-]+)"`)
)

func parseInfomation(buff []byte) (title, url string) {
	res := regexChannelTitle.FindSubmatch(buff)
	if len(res) == 2 {
		title = string(res[1])
	}
	if strings.HasPrefix(title, "=") ||
		strings.HasPrefix(title, "+") ||
		strings.HasPrefix(title, "-") {
		title = "'" + title
	}
	res = regexChannelURL.FindSubmatch(buff)
	if len(res) == 3 {
		url = fmt.Sprintf("https://www.youtube.com/%s%s", string(res[1]), string(res[2]))
	}
	return
}
