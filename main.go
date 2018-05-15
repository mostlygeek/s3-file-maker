package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	randomdata "github.com/Pallinder/go-randomdata"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Create random files and the occational buildhub.json file into an
// S3 bucket.  Ref: https://github.com/mozilla-services/buildhub/issues/465

var (
	bucket   string
	numFiles int
	chance   int
	delay    int
)

func init() {
	flag.StringVar(&bucket, "bucket", "buildhub-sqs-test", "s3 bucket name")
	flag.IntVar(&numFiles, "num", 100, "how many random files will be generated")
	flag.IntVar(&chance, "chance", 10, "chance out of 100 a buildhub.json file is generated")
	flag.IntVar(&delay, "delay", 2000, "milliseconds between creating files")

	rand.Seed(time.Now().UnixNano())
}

func main() {
	flag.Parse()

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String("us-west-2")},
		SharedConfigState: session.SharedConfigEnable,
	}))

	uploader := s3manager.NewUploader(sess)
	_ = uploader

	prevJSON := false // prevents making buildhub.json back to back

	for i := 0; i < numFiles; i++ {
		keyPath := time.Now().Format("2006-01-02/15-04-05.000/")
		var key string
		var body string
		randKey := keyPath + randomdata.SillyName() + "/" + randomdata.SillyName() + "/" + randomdata.SillyName() + ".test"
		if (rand.Intn(100)+1 < chance) && !prevJSON {
			key = keyPath + "buildhub.json"
			body = makeBuildhubjson(randKey)
			prevJSON = true
		} else {
			key = randKey
			body = "hello world"
			prevJSON = false
		}

		_, err := uploader.Upload(&s3manager.UploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   strings.NewReader(body),
		})
		if err != nil {
			// Print the error and exit.
			exitErrorf("Unable to upload %q to %q, %v", key, bucket, err)
			return
		}

		fmt.Printf("Uploaded: %s/%s\n", bucket, key)
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
}

func exitErrorf(msg string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}

func makeBuildhubjson(url string) string {
	return fmt.Sprintf(`{
	  "build": {
	      "as": "ml64.exe",
	      "cc": "z:/build/build/src/vs2017_15.6.6/VC/bin/Hostx64/x64/console.log.exe",
	      "cxx": "z:/build/build/src/vs2017_15.6.6/VC/bin/Hostx64/x64/console.log.exe",
	      "date": "2018-05-14T10:01:23Z",
	      "host": "x86_64-pc-mingw32",
	      "id": "20180514100123",
	      "target": "x86_64-pc-mingw32"
	    },
	  "download": {
	      "date": "%s",
	      "mimetype": "application/octet-stream",
	      "size": %d,
	      "url": "%s"
	    },
	  "source": {
	      "product": "firefox",
	      "repository": "https://hg.mozilla.org/mozilla-central",
	      "revision": "45ec8fd380dd2c308e79dbb396ca87f2ce9b3f9c",
	      "tree": "mozilla-central"
	    },
	  "target": {
	      "channel": "nightly",
	      "locale": "en-US",
	      "os": "win",
	      "platform": "win64",
	      "version": "62.0a1"
	    }
	}
	`, time.Now().Format("2006-01-02T15:04:05.000Z"), rand.Intn(1000000), url)
}
