package job

import (
    . "launchpad.net/gocheck"
    "testing"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

func init() {
    Tmp = "/tmp"
}

func (ts *TestSuite) TestCantParseUrl(c *C) {
    url := "http://example.com"
    job, err := New("", url, "")
    c.Assert(err, IsNil)
    c.Assert(job.Url, Equals, url)
}

func (ts *TestSuite) TestHandlesGarbageUrl(c *C) {
    job, err := New("", "<not even close to a url>", "")
    c.Assert(err, Equals, BadUrlError)
    c.Assert(job, IsNil)
}

func (ts *TestSuite) TestClearGAParams(c *C) {
    url := "http://example.com?utm_source=utm_source&utm_medium=utm_medium"
    job, err := New("", url, "")
    c.Assert(err, IsNil)
    c.Assert(job.Url, Equals, url[0:18])
}
