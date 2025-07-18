package search

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Client defines the interface for interacting with the Watchman service.
type Client interface {
	// ListInfo retrieves the available data lists from a Watchman service.
	ListInfo(ctx context.Context) (ListInfoResponse, error)

	// SearchByEntity performs a search for entities (e.g., individuals, businesses) based
	// on the provided query and search options.
	//
	// The query parameter specifies the entity fields to search for - populated fields are compared.
	//
	// The opts parameter allows customization of the search behavior, such as filters or limits.
	//
	// The method returns a SearchResponse containing the search results or an error if the
	// request fails.
	//
	// Example:
	//   ctx := context.Background()
	//   query := search.Entity[search.Value]{...}
	//   opts := SearchOpts{Limit: 10}
	//
	//   resp, err := client.SearchByEntity(ctx, query, opts)
	//   if err != nil {
	//       log.Fatal(err)
	//   }
	SearchByEntity(ctx context.Context, entity Entity[Value], opts SearchOpts) (SearchResponse, error)

	// IngestFile uploads a file to the Watchman service for updating a data set.
	//
	// The fileType parameter specifies the type (dataset name) of file being uploaded (e.g., fincen-person).
	// This aligns with a schema defined under Watchman.Ingest.Files in the yaml config.
	// Currently only CSV files are supported.
	//
	// The file parameter is an io.Reader containing the file data. The method returns an
	// IngestFileResponse with details about the ingestion process or an error if the request fails.
	//
	// Example:
	//   file, err := os.Open("sanctions.csv")
	//   if err != nil {
	//       log.Fatal(err)
	//   }
	//   defer file.Close()
	//
	//   resp, err := client.IngestFile(ctx, "fincen-person", file)
	//   if err != nil {
	//       log.Fatal(err)
	//   }
	IngestFile(ctx context.Context, fileType string, file io.Reader) (IngestFileResponse, error)
}

func NewClient(httpClient *http.Client, baseAddress string) Client {
	httpClient = cmp.Or(httpClient, &http.Client{
		Timeout: 20 * time.Second,
	})

	retryableclient := retryablehttp.NewClient()
	retryableclient.HTTPClient = httpClient
	retryableclient.Logger = nil // disable logging

	return &client{
		client:      retryableclient,
		baseAddress: baseAddress,
	}
}

type client struct {
	client      *retryablehttp.Client
	baseAddress string
}

type ListInfoResponse struct {
	Lists      map[string]int    `json:"lists"`
	ListHashes map[string]string `json:"listHashes"`

	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`

	// Version is the version of Watchman that returned results.
	Version string `json:"version"`
}

// ListInfo retrieves the available data lists from a Watchman service.
func (c *client) ListInfo(ctx context.Context) (ListInfoResponse, error) {
	addr := c.baseAddress + "/v2/listinfo"

	var out ListInfoResponse
	req, err := retryablehttp.NewRequest("GET", addr, nil)
	if err != nil {
		return out, fmt.Errorf("creating listinfo request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return out, fmt.Errorf("listinfo GET: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return out, fmt.Errorf("decoding listinfo response: %w", err)
	}
	return out, nil
}

type SearchResponse struct {
	Query Entity[Value] `json:"query"`

	Entities []SearchedEntity[Value] `json:"entities"`
}

func (s *SearchResponse) UnmarshalJSON(data []byte) error {
	var aux struct {
		Entities []SearchedEntity[Value] `json:"entities"`
		Error    string                  `json:"error"`
	}
	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}
	if aux.Error != "" {
		return errors.New(aux.Error)
	}

	s.Entities = aux.Entities

	return nil
}

type SearchOpts struct {
	Limit    int
	MinMatch float64
	Debug    bool
}

// SearchByEntity searches for entities (e.g., individuals, businesses) using the provided query fields and
// search options via a GET request to /v2/search.
//
// The entity parameter specifies the fields to search for, where populated fields are matched against the Watchman instance.
// The opts parameter customizes the search with options like limit (result count), minimum match score (MinMatch), or debug mode.
//
// The method returns a SearchResponse containing the query and matching entities or an error if the request, URL construction, or JSON decoding fails.
func (c *client) SearchByEntity(ctx context.Context, entity Entity[Value], opts SearchOpts) (SearchResponse, error) {
	var out SearchResponse

	// Build the URL
	addr, err := url.Parse(c.baseAddress + "/v2/search")
	if err != nil {
		return out, fmt.Errorf("problem creating baseAddress: %w", err)
	}

	// Set query parameters
	addr.RawQuery = BuildQueryParameters(SetSearchOpts(addr.Query(), opts), entity).Encode()

	// Make the request
	req, err := retryablehttp.NewRequest("GET", addr.String(), nil)
	if err != nil {
		return out, fmt.Errorf("creating search request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return out, fmt.Errorf("search by entity: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return out, fmt.Errorf("decoding search by entity response: %w", err)
	}
	return out, nil
}

func SetSearchOpts(q url.Values, opts SearchOpts) url.Values {
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.MinMatch > 0.00 {
		q.Set("minMatch", fmt.Sprintf("%.2f", opts.MinMatch))
	}
	if opts.Debug {
		q.Set("debug", "yes")
	}

	return q
}

func BuildQueryParameters(q url.Values, entity Entity[Value]) url.Values {
	q.Set("type", string(entity.Type))

	if src := string(entity.Source); src != "" {
		q.Set("source", src)
	}

	if entity.Name != "" {
		q.Set("name", entity.Name)
	}

	// Person, Business, Organization, Aircraft, Vessel
	if entity.Person != nil {
		setPersonParameters(q, entity)
	}
	if entity.Business != nil {
		setBusinessParameters(q, entity)
	}
	if entity.Organization != nil {
		setOrganizationParameters(q, entity)
	}
	if entity.Aircraft != nil {
		setAircraftParameters(q, entity)
	}
	if entity.Vessel != nil {
		setVesselParameters(q, entity)
	}

	// ContactInfo, Addresses, CryptoAddresses, etc
	setContactInfo(q, entity.Contact)
	setAddresses(q, entity.Addresses)
	setCryptoAddresses(q, entity.CryptoAddresses)

	return q
}

const (
	yyyymmdd = "2006-01-02"
)

func setPersonParameters(q url.Values, entity Entity[Value]) {
	if entity.Person == nil {
		return
	}

	if entity.Person.Name != "" {
		q.Set("name", entity.Person.Name) // replaces what was set
	}
	for _, alt := range entity.Person.AltNames {
		q.Add("altNames", alt)
	}
	if g := string(entity.Person.Gender); g != "" {
		q.Set("gender", g)
	}
	if entity.Person.BirthDate != nil {
		q.Set("birthDate", entity.Person.BirthDate.Format(yyyymmdd))
	}
	for _, title := range entity.Person.Titles {
		q.Add("titles", title)
	}

	setGovernmentIDs(q, entity.Person.GovernmentIDs)
}

func setBusinessParameters(q url.Values, entity Entity[Value]) {
	if entity.Business == nil {
		return
	}

	if entity.Business.Name != "" {
		q.Set("name", entity.Business.Name) // replaces what was set
	}
	for _, alt := range entity.Business.AltNames {
		q.Add("altNames", alt)
	}
	if entity.Business.Created != nil {
		q.Set("created", entity.Business.Created.Format(yyyymmdd))
	}

	setGovernmentIDs(q, entity.Business.GovernmentIDs)
}

func setOrganizationParameters(q url.Values, entity Entity[Value]) {
	if entity.Organization == nil {
		return
	}

	if entity.Organization.Name != "" {
		q.Set("name", entity.Organization.Name) // replaces what was set
	}
	for _, alt := range entity.Organization.AltNames {
		q.Add("altNames", alt)
	}
	if entity.Organization.Created != nil {
		q.Set("created", entity.Organization.Created.Format(yyyymmdd))
	}

	setGovernmentIDs(q, entity.Organization.GovernmentIDs)
}

func setAircraftParameters(q url.Values, entity Entity[Value]) {
	if entity.Aircraft == nil {
		return
	}

	if entity.Aircraft.Name != "" {
		q.Set("name", entity.Aircraft.Name) // replaces what was set
	}
	for _, alt := range entity.Aircraft.AltNames {
		q.Add("altNames", alt)
	}
	if t := string(entity.Aircraft.Type); t != "" {
		q.Set("aircraftType", t)
	}
	if entity.Aircraft.Flag != "" {
		q.Set("flag", entity.Aircraft.Flag)
	}
	if entity.Aircraft.Built != nil {
		q.Set("built", entity.Aircraft.Built.Format(yyyymmdd))
	}
	if entity.Aircraft.ICAOCode != "" {
		q.Set("icaoCode", entity.Aircraft.ICAOCode)
	}
	if entity.Aircraft.Model != "" {
		q.Set("model", entity.Aircraft.Model)
	}
	if entity.Aircraft.SerialNumber != "" {
		q.Set("serialNumber", entity.Aircraft.SerialNumber)
	}
}

func setVesselParameters(q url.Values, entity Entity[Value]) {
	if entity.Vessel == nil {
		return
	}

	if entity.Vessel.Name != "" {
		q.Set("name", entity.Vessel.Name) // replaces what was set
	}
	for _, alt := range entity.Vessel.AltNames {
		q.Add("altNames", alt)
	}
	if entity.Vessel.IMONumber != "" {
		q.Set("imoNumber", entity.Vessel.IMONumber)
	}
	if t := string(entity.Vessel.Type); t != "" {
		q.Set("vesselType", t)
	}
	if entity.Vessel.Flag != "" {
		q.Set("flag", entity.Vessel.Flag)
	}
	if entity.Vessel.Built != nil {
		q.Set("built", entity.Vessel.Built.Format(yyyymmdd))
	}
	if entity.Vessel.Model != "" {
		q.Set("model", entity.Vessel.Model)
	}
	if entity.Vessel.Tonnage > 0 {
		q.Set("tonnage", strconv.Itoa(entity.Vessel.Tonnage))
	}
	if entity.Vessel.MMSI != "" {
		q.Set("mmsi", entity.Vessel.MMSI)
	}
	if entity.Vessel.CallSign != "" {
		q.Set("callSign", entity.Vessel.CallSign)
	}
	// TODO(adam): GrossRegisteredTonnage
	if entity.Vessel.Owner != "" {
		q.Set("owner", entity.Vessel.Owner)
	}
}

func setGovernmentIDs(q url.Values, ids []GovernmentID) {
	for _, id := range ids {
		key := fmt.Sprintf("gov_%s", string(id.Type))
		value := fmt.Sprintf("%s:%s", id.Country, id.Identifier)

		q.Add(key, value)
	}
}

func setContactInfo(q url.Values, info ContactInfo) {
	for _, email := range info.EmailAddresses {
		q.Add("emailAddress", email)
	}
	for _, phone := range info.PhoneNumbers {
		q.Add("phoneNumber", phone)
	}
	for _, fax := range info.FaxNumbers {
		q.Add("faxNumber", fax)
	}
	for _, website := range info.Websites {
		q.Add("website", website)
	}
}

func setAddresses(q url.Values, addresses []Address) {
	for _, addr := range addresses {
		q.Add("address", addr.Format())
	}
}

func setCryptoAddresses(q url.Values, cryptoAddresses []CryptoAddress) {
	for _, addr := range cryptoAddresses {
		q.Add("cryptoAddress", fmt.Sprintf("%s:%s", addr.Currency, addr.Address))
	}
}

type IngestFileResponse struct {
	FileType string          `json:"fileType"`
	Entities []Entity[Value] `json:"entities"`
}

// IngestFile uploads a CSV file to the Watchman service to update a dataset via a POST request to /v2/ingest/{fileType}.
//
// The fileType parameter specifies the dataset name (e.g., "fincen-person") as defined in the
// Watchman.Ingest.Files YAML configuration. The file parameter is an io.Reader containing the
// CSV file data.
//
// The method returns an IngestFileResponse with the file type and ingested entities
// or an error if the request, URL construction, or JSON decoding fails.
func (c *client) IngestFile(ctx context.Context, fileType string, file io.Reader) (IngestFileResponse, error) {
	var out IngestFileResponse

	// Build the URL
	addr, err := url.Parse(c.baseAddress + fmt.Sprintf("/v2/ingest/%s", fileType))
	if err != nil {
		return out, fmt.Errorf("problem creating baseAddress: %w", err)
	}

	// Make the request
	req, err := retryablehttp.NewRequest("POST", addr.String(), file)
	if err != nil {
		return out, fmt.Errorf("creating ingest file request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return out, fmt.Errorf("ingest file: %w", err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	err = json.NewDecoder(resp.Body).Decode(&out)
	if err != nil {
		return out, fmt.Errorf("decoding ingest file response: %w", err)
	}
	return out, nil
}
