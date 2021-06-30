package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

type consumer struct {
	statusCodeOnly bool
	bar            *ProgressBar
	log            *logrus.Logger
	exclude        []string
	errorsLine     []string
	errorsLineFile *os.File
}

func NewConsumer(statusCodeOnly bool, bar *ProgressBar, log *logrus.Logger, exclude []string, file *os.File) Consumer {
	t := make([]string, 0)
	return &consumer{
		statusCodeOnly: statusCodeOnly,
		bar:            bar,
		log:            log,
		exclude:        exclude,
		errorsLine:     t,
		errorsLineFile: file,
	}
}

func (c *consumer) Consume(val HostsPair) {
	if val.HasErrors() {
		c.bar.IncrementError()
		for _, v := range val.Errors {
			c.log.Errorln(v)
		}
		return
	}

	if val.EqualStatusCode() {
		switch {
		case val.Left.StatusCode >= 200 && val.Left.StatusCode <= 299:
			c.bar.IncrementStatus2XX()
		case val.Left.StatusCode == 401:
			c.bar.IncrementStatus401()
		case val.Left.StatusCode != 401 && val.Left.StatusCode >= 400 && val.Left.StatusCode <= 499:
			c.bar.IncrementStatus4XX()
			c.log.Warnf("error %d in url %s", val.Left.StatusCode, val.RelURL)
		}
	}

	if val.EqualStatusCode() && c.statusCodeOnly {
		c.bar.IncrementOk()
		return
	}

	if !val.EqualStatusCode() {
		c.bar.IncrementError()
		c.log.Warnf("found status code diff: url %s, %s: %d - %s: %d",
			val.RelURL, val.Left.URL.Host, val.Left.StatusCode, val.Right.URL.Host, val.Right.StatusCode)
		return
	}

	leftJSON, err := unmarshal(val.Left.Body)
	if err != nil {
		c.bar.IncrementError()
		c.log.Errorf("could not unmarshal json: url %s: %v", val.RelURL, err)
		return
	}

	rightJSON, err := unmarshal(val.Right.Body)
	if err != nil {
		c.bar.IncrementError()
		c.log.Errorf("could not unmarshal json: url %s: %v", val.RelURL, err)
		return
	}

	for _, exclude := range c.exclude {
		Remove(leftJSON, exclude)
		Remove(rightJSON, exclude)
	}

	if !Equal(leftJSON, rightJSON) {
		c.bar.IncrementError()
		c.log.Warnf("found json diff: url %s", val.RelURL)
		c.errorsLine = append(c.errorsLine, val.RelURL)
		c.errorsLineFile.Write([]byte(val.RelURL))
		c.errorsLineFile.Write([]byte("\n"))
		return
	}

	c.bar.IncrementOk()
}

func unmarshal(b []byte) (interface{}, error) {
	j, err := Unmarshal(b)
	if err != nil {
		return nil, err
	}

	return j, nil
}
