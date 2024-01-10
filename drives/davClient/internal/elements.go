package internal

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const Namespace = "DAV:"

var (
	ResourceTypeName     = xml.Name{Namespace, "resourcetype"}
	DisplayNameName      = xml.Name{Namespace, "displayname"}
	GetContentLengthName = xml.Name{Namespace, "getcontentlength"}
	GetContentTypeName   = xml.Name{Namespace, "getcontenttype"}
	GetLastModifiedName  = xml.Name{Namespace, "getlastmodified"}
	GetETagName          = xml.Name{Namespace, "getetag"}

	CurrentUserPrincipalName = xml.Name{Namespace, "current-user-principal"}
)

type Status struct {
	Code int
	Text string
}

func (s *Status) MarshalText() ([]byte, error) {
	text := s.Text
	if text == "" {
		text = http.StatusText(s.Code)
	}
	return []byte(fmt.Sprintf("HTTP/1.1 %v %v", s.Code, text)), nil
}

func (s *Status) UnmarshalText(b []byte) error {
	if len(b) == 0 {
		return nil
	}

	parts := strings.SplitN(string(b), " ", 3)
	if len(parts) != 3 {
		return fmt.Errorf("webdav: invalid HTTP status %q: expected 3 fields", s)
	}
	code, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("webdav: invalid HTTP status %q: failed to parse code: %v", s, err)
	}

	s.Code = code
	s.Text = parts[2]
	return nil
}

func (s *Status) Err() error {
	if s == nil {
		return nil
	}

	// TODO: handle 2xx, 3xx
	if s.Code != http.StatusOK {
		return &HTTPError{Code: s.Code}
	}
	return nil
}

type Href url.URL

func (h *Href) String() string {
	u := (*url.URL)(h)
	return u.String()
}

func (h *Href) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *Href) UnmarshalText(b []byte) error {
	u, err := url.Parse(string(b))
	if err != nil {
		return err
	}
	*h = Href(*u)
	return nil
}

// https://tools.ietf.org/html/rfc4918#section-14.16
type MultiStatus struct {
	XMLName             xml.Name   `xml:"DAV: multistatus"`
	Responses           []Response `xml:"response"`
	ResponseDescription string     `xml:"responsedescription,omitempty"`
	SyncToken           string     `xml:"sync-token,omitempty"`
}

func NewMultiStatus(resps ...Response) *MultiStatus {
	return &MultiStatus{Responses: resps}
}

// https://tools.ietf.org/html/rfc4918#section-14.24
type Response struct {
	XMLName             xml.Name   `xml:"DAV: response"`
	Hrefs               []Href     `xml:"href"`
	PropStats           []PropStat `xml:"propstat,omitempty"`
	ResponseDescription string     `xml:"responsedescription,omitempty"`
	Status              *Status    `xml:"status,omitempty"`
	Error               *Error     `xml:"error,omitempty"`
	Location            *Location  `xml:"location,omitempty"`
}

func NewOKResponse(path string) *Response {
	href := Href{Path: path}
	return &Response{
		Hrefs:  []Href{href},
		Status: &Status{Code: http.StatusOK},
	}
}

func NewErrorResponse(path string, err error) *Response {
	code := http.StatusInternalServerError
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		code = httpErr.Code
	}

	var errElt *Error
	errors.As(err, &errElt)

	href := Href{Path: path}
	return &Response{
		Hrefs:               []Href{href},
		Status:              &Status{Code: code},
		ResponseDescription: err.Error(),
		Error:               errElt,
	}
}

func (resp *Response) Err() error {
	if resp.Status == nil || resp.Status.Code/100 == 2 {
		return nil
	}

	var err error
	if resp.Error != nil {
		err = resp.Error
	}
	if resp.ResponseDescription != "" {
		if err != nil {
			err = fmt.Errorf("%v (%w)", resp.ResponseDescription, err)
		} else {
			err = fmt.Errorf("%v", resp.ResponseDescription)
		}
	}

	return &HTTPError{
		Code: resp.Status.Code,
		Err:  err,
	}
}

func (resp *Response) Path() (string, error) {
	err := resp.Err()
	var path string
	if len(resp.Hrefs) == 1 {
		path = resp.Hrefs[0].Path
	} else if err == nil {
		err = fmt.Errorf("webdav: malformed response: expected exactly one href element, got %v", len(resp.Hrefs))
	}
	return path, err
}

func (resp *Response) DecodeProp(values ...interface{}) error {
	for _, v := range values {
		// TODO wrap errors with more context (XML name)
		name, err := valueXMLName(v)
		if err != nil {
			return err
		}
		if err := resp.Err(); err != nil {
			return newPropError(name, err)
		}
		for _, propstat := range resp.PropStats {
			raw := propstat.Prop.Get(name)
			if raw == nil {
				continue
			}
			if err := propstat.Status.Err(); err != nil {
				return newPropError(name, err)
			}
			if err := raw.Decode(v); err != nil {
				return newPropError(name, err)
			}
			return nil
		}
		return newPropError(name, &HTTPError{
			Code: http.StatusNotFound,
			Err:  fmt.Errorf("missing property"),
		})
	}

	return nil
}

func newPropError(name xml.Name, err error) error {
	return fmt.Errorf("property <%v %v>: %w", name.Space, name.Local, err)
}

func (resp *Response) EncodeProp(code int, v interface{}) error {
	raw, err := EncodeRawXMLElement(v)
	if err != nil {
		return err
	}

	for i := range resp.PropStats {
		propstat := &resp.PropStats[i]
		if propstat.Status.Code == code {
			propstat.Prop.Raw = append(propstat.Prop.Raw, *raw)
			return nil
		}
	}

	resp.PropStats = append(resp.PropStats, PropStat{
		Status: Status{Code: code},
		Prop:   Prop{Raw: []RawXMLValue{*raw}},
	})
	return nil
}

// https://tools.ietf.org/html/rfc4918#section-14.9
type Location struct {
	XMLName xml.Name `xml:"DAV: location"`
	Href    Href     `xml:"href"`
}

// https://tools.ietf.org/html/rfc4918#section-14.22
type PropStat struct {
	XMLName             xml.Name `xml:"DAV: propstat"`
	Prop                Prop     `xml:"prop"`
	Status              Status   `xml:"status"`
	ResponseDescription string   `xml:"responsedescription,omitempty"`
	Error               *Error   `xml:"error,omitempty"`
}

// https://tools.ietf.org/html/rfc4918#section-14.18
type Prop struct {
	XMLName xml.Name      `xml:"DAV: prop"`
	Raw     []RawXMLValue `xml:",any"`
}

func EncodeProp(values ...interface{}) (*Prop, error) {
	l := make([]RawXMLValue, len(values))
	for i, v := range values {
		raw, err := EncodeRawXMLElement(v)
		if err != nil {
			return nil, err
		}
		l[i] = *raw
	}
	return &Prop{Raw: l}, nil
}

func (p *Prop) Get(name xml.Name) *RawXMLValue {
	for i := range p.Raw {
		raw := &p.Raw[i]
		if n, ok := raw.XMLName(); ok && name == n {
			return raw
		}
	}
	return nil
}

func (p *Prop) Decode(v interface{}) error {
	name, err := valueXMLName(v)
	if err != nil {
		return err
	}

	raw := p.Get(name)
	if raw == nil {
		return HTTPErrorf(http.StatusNotFound, "missing property %s", name)
	}

	return raw.Decode(v)
}

// https://tools.ietf.org/html/rfc4918#section-14.20
type PropFind struct {
	XMLName  xml.Name  `xml:"DAV: propfind"`
	Prop     *Prop     `xml:"prop,omitempty"`
	AllProp  *struct{} `xml:"allprop,omitempty"`
	Include  *Include  `xml:"include,omitempty"`
	PropName *struct{} `xml:"propname,omitempty"`
}

func xmlNamesToRaw(names []xml.Name) []RawXMLValue {
	l := make([]RawXMLValue, len(names))
	for i, name := range names {
		l[i] = *NewRawXMLElement(name, nil, nil)
	}
	return l
}

func NewPropNamePropFind(names ...xml.Name) *PropFind {
	return &PropFind{Prop: &Prop{Raw: xmlNamesToRaw(names)}}
}

// https://tools.ietf.org/html/rfc4918#section-14.8
type Include struct {
	XMLName xml.Name      `xml:"DAV: include"`
	Raw     []RawXMLValue `xml:",any"`
}

// https://tools.ietf.org/html/rfc4918#section-15.9
type ResourceType struct {
	XMLName xml.Name      `xml:"DAV: resourcetype"`
	Raw     []RawXMLValue `xml:",any"`
}

func NewResourceType(names ...xml.Name) *ResourceType {
	return &ResourceType{Raw: xmlNamesToRaw(names)}
}

func (t *ResourceType) Is(name xml.Name) bool {
	for _, raw := range t.Raw {
		if n, ok := raw.XMLName(); ok && name == n {
			return true
		}
	}
	return false
}

var CollectionName = xml.Name{Namespace, "collection"}

// https://tools.ietf.org/html/rfc4918#section-15.4
type GetContentLength struct {
	XMLName xml.Name `xml:"DAV: getcontentlength"`
	Length  int64    `xml:",chardata"`
}

// https://tools.ietf.org/html/rfc4918#section-15.5
type GetContentType struct {
	XMLName xml.Name `xml:"DAV: getcontenttype"`
	Type    string   `xml:",chardata"`
}

type Time time.Time

func (t *Time) UnmarshalText(b []byte) error {
	tt, err := http.ParseTime(string(b))
	if err != nil {
		return err
	}
	*t = Time(tt)
	return nil
}

func (t *Time) MarshalText() ([]byte, error) {
	s := time.Time(*t).UTC().Format(http.TimeFormat)
	return []byte(s), nil
}

// https://tools.ietf.org/html/rfc4918#section-15.7
type GetLastModified struct {
	XMLName      xml.Name `xml:"DAV: getlastmodified"`
	LastModified Time     `xml:",chardata"`
}

// https://tools.ietf.org/html/rfc4918#section-15.6
type GetETag struct {
	XMLName xml.Name `xml:"DAV: getetag"`
	ETag    ETag     `xml:",chardata"`
}

type ETag string

func (etag *ETag) UnmarshalText(b []byte) error {
	s, err := strconv.Unquote(string(b))
	if err != nil {
		return fmt.Errorf("webdav: failed to unquote ETag: %v", err)
	}
	*etag = ETag(s)
	return nil
}

func (etag ETag) MarshalText() ([]byte, error) {
	return []byte(etag.String()), nil
}

func (etag ETag) String() string {
	return fmt.Sprintf("%q", string(etag))
}

// https://tools.ietf.org/html/rfc4918#section-14.5
type Error struct {
	XMLName xml.Name      `xml:"DAV: error"`
	Raw     []RawXMLValue `xml:",any"`
}

func (err *Error) Error() string {
	b, _ := xml.Marshal(err)
	return string(b)
}

// https://tools.ietf.org/html/rfc4918#section-15.2
type DisplayName struct {
	XMLName xml.Name `xml:"DAV: displayname"`
	Name    string   `xml:",chardata"`
}

// https://tools.ietf.org/html/rfc5397#section-3
type CurrentUserPrincipal struct {
	XMLName         xml.Name  `xml:"DAV: current-user-principal"`
	Href            Href      `xml:"href,omitempty"`
	Unauthenticated *struct{} `xml:"unauthenticated,omitempty"`
}

// https://tools.ietf.org/html/rfc4918#section-14.19
type PropertyUpdate struct {
	XMLName xml.Name `xml:"DAV: propertyupdate"`
	Remove  []Remove `xml:"remove"`
	Set     []Set    `xml:"set"`
}

// https://tools.ietf.org/html/rfc4918#section-14.23
type Remove struct {
	XMLName xml.Name `xml:"DAV: remove"`
	Prop    Prop     `xml:"prop"`
}

// https://tools.ietf.org/html/rfc4918#section-14.26
type Set struct {
	XMLName xml.Name `xml:"DAV: set"`
	Prop    Prop     `xml:"prop"`
}

// https://tools.ietf.org/html/rfc6578#section-6.1
type SyncCollectionQuery struct {
	XMLName   xml.Name `xml:"DAV: sync-collection"`
	SyncToken string   `xml:"sync-token"`
	Limit     *Limit   `xml:"limit,omitempty"`
	SyncLevel string   `xml:"sync-level"`
	Prop      *Prop    `xml:"prop"`
}

// https://tools.ietf.org/html/rfc5323#section-5.17
type Limit struct {
	XMLName  xml.Name `xml:"DAV: limit"`
	NResults uint     `xml:"nresults"`
}
