package crawler

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/infralens/infralens/internal/client"
	"github.com/infralens/infralens/internal/model"
	"github.com/infralens/infralens/internal/store"
)

const (
	workers      = 5
	rateDelay    = 300 * time.Millisecond // delay between requests per worker
)

type Crawler struct {
	client *client.Client
	store  *store.Store
}

func New(c *client.Client, s *store.Store) *Crawler {
	return &Crawler{client: c, store: s}
}

func (cr *Crawler) Run(ctx context.Context, startID, endID int) (processed int, failed int) {
	jobs := make(chan int, workers*2)
	var wg sync.WaitGroup
	var processedAtomic, failedAtomic atomic.Int64

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for projectID := range jobs {
				if err := cr.crawlOne(ctx, projectID); err != nil {
					if errors.Is(err, client.ErrNotFound) {
						log.Printf("[SKIP] project %d: not found", projectID)
					} else {
						log.Printf("[ERR]  project %d: %v", projectID, err)
						failedAtomic.Add(1)
					}
				} else {
					processedAtomic.Add(1)
					log.Printf("[OK]   project %d (total: %d)", projectID, processedAtomic.Load())
				}
				time.Sleep(rateDelay)
			}
		}()
	}

	for id := startID; id <= endID; id++ {
		select {
		case <-ctx.Done():
			break
		case jobs <- id:
		}
	}
	close(jobs)
	wg.Wait()

	processed = int(processedAtomic.Load())
	failed = int(failedAtomic.Load())
	log.Printf("Done. processed=%d failed=%d", processed, failed)
	return
}

func (cr *Crawler) crawlOne(ctx context.Context, projectID int) error {
	// Step 1: fetch general details (gives us userProfileId)
	general, err := cr.client.GetProjectGeneral(ctx, projectID)
	if err != nil {
		return err
	}

	// Build snapshot checksum from general details to detect future changes
	generalJSON, _ := json.Marshal(general)
	checksum := fmt.Sprintf("%x", md5.Sum(generalJSON))

	// Step 2: fetch everything else in parallel
	type result[T any] struct {
		val T
		err error
	}

	var (
		promoterCh      = make(chan result[*model.PromoterGeneral], 1)
		promoterAssocCh = make(chan result[*model.PromoterAssociation], 1)
		promoterAddrCh  = make(chan result[*model.PromoterAddress], 1)
		addrCh          = make(chan result[*model.ProjectAddress], 1)
		profCh          = make(chan result[[]model.Professional], 1)
		agentCh         = make(chan result[[]model.Agent], 1)
	)

	go func() {
		v, e := cr.client.GetPromoterGeneral(ctx, general.UserProfileID, projectID)
		promoterCh <- result[*model.PromoterGeneral]{v, e}
	}()
	go func() {
		v, e := cr.client.GetPromoterAssociation(ctx, projectID)
		promoterAssocCh <- result[*model.PromoterAssociation]{v, e}
	}()
	go func() {
		v, e := cr.client.GetPromoterAddress(ctx, general.UserProfileID, projectID)
		promoterAddrCh <- result[*model.PromoterAddress]{v, e}
	}()
	go func() {
		v, e := cr.client.GetProjectAddress(ctx, projectID)
		addrCh <- result[*model.ProjectAddress]{v, e}
	}()
	go func() {
		v, e := cr.client.GetProfessionals(ctx, projectID, "")
		profCh <- result[[]model.Professional]{v, e}
	}()
	go func() {
		v, e := cr.client.GetAgents(ctx, projectID)
		agentCh <- result[[]model.Agent]{v, e}
	}()

	promoterResult      := <-promoterCh
	promoterAssocResult := <-promoterAssocCh
	promoterAddrResult  := <-promoterAddrCh
	addrResult          := <-addrCh
	profResult          := <-profCh
	agentResult         := <-agentCh

	// Step 3: persist promoter first (projects.promoter_id FK)
	// fetchPromoterGeneralDetails often returns empty name — fall back to association
	var promoterDBID *int
	if promoterResult.err == nil && promoterResult.val != nil {
		p := promoterResult.val
		if p.PromoterName == "" && promoterAssocResult.val != nil {
			p.PromoterName = promoterAssocResult.val.PromoterName
		}
		raw, _ := json.Marshal(p)
		dbID, err := cr.store.UpsertPromoter(ctx, &model.Promoter{
			UserProfileID: p.UserProfileID,
			Name:          p.PromoterName,
			Pan:           p.Pan,
			Gstin:         p.Gstin,
			PromoterType:  p.PromoterType,
			RawJSON:       raw,
		})
		if err != nil {
			log.Printf("[WARN] project %d promoter upsert: %v", projectID, err)
		} else {
			promoterDBID = &dbID
		}
	}

	// Step 4: persist project
	proj := buildProject(general, promoterDBID, generalJSON)
	projectDBID, err := cr.store.UpsertProject(ctx, proj)
	if err != nil {
		return fmt.Errorf("upsert project: %w", err)
	}

	// Step 5a: persist project address
	if addrResult.err == nil && addrResult.val != nil {
		a := addrResult.val
		raw, _ := json.Marshal(a)
		_ = cr.store.InsertAddress(ctx, &model.Address{
			EntityType: "project",
			EntityID:   projectDBID,
			Line1:      a.AddressLine,
			Line2:      a.Street,
			City:       a.Locality,
			District:   a.DistrictName,
			State:      a.StateName,
			Pincode:    a.PinCode,
			RawJSON:    raw,
		})
	}

	// Step 5b: persist promoter address (linked via promoter DB id)
	if promoterAddrResult.err == nil && promoterAddrResult.val != nil && promoterDBID != nil {
		a := promoterAddrResult.val
		raw, _ := json.Marshal(a)
		line1 := a.BuildingName
		if a.UnitNumber != "" {
			line1 = a.UnitNumber + ", " + a.BuildingName
		}
		_ = cr.store.InsertAddress(ctx, &model.Address{
			EntityType: "promoter",
			EntityID:   *promoterDBID,
			Line1:      line1,
			Line2:      a.StreetName,
			City:       a.Locality,
			District:   a.DistrictName,
			State:      a.StateName,
			Pincode:    a.PinCode,
			RawJSON:    raw,
		})
	}

	// Step 6: persist contacts (professionals + agents)
	var contacts []model.Contact
	if profResult.err == nil {
		for _, p := range profResult.val {
			raw, _ := json.Marshal(p)
			contacts = append(contacts, model.Contact{
				ProjectID: projectDBID,
				Role:      p.ProfessionalType,
				Name:      p.ProfessionalName,
				Phone:     p.MobileNumber,
				Email:     p.EmailID,
				RawJSON:   raw,
			})
		}
	}
	if agentResult.err == nil {
		for _, a := range agentResult.val {
			raw, _ := json.Marshal(a)
			contacts = append(contacts, model.Contact{
				ProjectID: projectDBID,
				Role:      "Agent",
				Name:      a.AgentName,
				Phone:     a.MobileNumber,
				Email:     a.EmailID,
				RawJSON:   raw,
			})
		}
	}
	if len(contacts) > 0 {
		_ = cr.store.InsertContacts(ctx, contacts)
	}

	// Step 7: change detection — compare new checksum against latest snapshot
	now := time.Now()
	latestSnap, err := cr.store.GetLatestSnapshot(ctx, projectDBID)
	if err != nil {
		log.Printf("[WARN] project %d get snapshot: %v", projectID, err)
	}

	if latestSnap == nil {
		// First time seeing this project — no diff, just snapshot
		log.Printf("[NEW]  project %d (%s)", projectID, general.ProjectName)
	} else if latestSnap.Checksum != checksum {
		// Something changed — diff field by field
		var prev model.ProjectGeneral
		if err := json.Unmarshal(latestSnap.RawJSON, &prev); err != nil {
			log.Printf("[WARN] project %d unmarshal prev snapshot: %v", projectID, err)
		} else {
			changes := diffGeneralFields(projectDBID, &prev, general, now)
			if len(changes) > 0 {
				if err := cr.store.InsertProjectChanges(ctx, changes); err != nil {
					log.Printf("[WARN] project %d insert changes: %v", projectID, err)
				} else {
					log.Printf("[DIFF] project %d: %d field(s) changed", projectID, len(changes))
					for _, c := range changes {
						log.Printf("       %s: %q → %q", c.FieldName, c.OldValue, c.NewValue)
					}
				}
			}
		}
	} else {
		log.Printf("[SAME] project %d — no changes", projectID)
	}

	// Always write a new snapshot to track crawl history
	_ = cr.store.InsertSnapshot(ctx, &model.ProjectSnapshot{
		ProjectID: projectDBID,
		FetchedAt: now,
		Checksum:  checksum,
		RawJSON:   generalJSON,
	})

	return nil
}

// trackedFields defines which fields we diff on every crawl.
// Add or remove fields here to control what gets tracked in project_changes.
var trackedFields = []struct {
	name string
	get  func(*model.ProjectGeneral) string
}{
	{"project_name", func(p *model.ProjectGeneral) string { return p.ProjectName }},
	{"project_status", func(p *model.ProjectGeneral) string { return p.ProjectStatusName }},
	{"project_current_status", func(p *model.ProjectGeneral) string { return p.ProjectCurrentStatus }},
	{"proposed_completion_date", func(p *model.ProjectGeneral) string { return p.ProjectProposeComplitionDate }},
	{"total_units", func(p *model.ProjectGeneral) string { return fmt.Sprintf("%d", p.TotalNumberOfUnits) }},
	{"total_sold_units", func(p *model.ProjectGeneral) string { return fmt.Sprintf("%d", p.TotalNumberOfSoldUnits) }},
	{"rera_registration_no", func(p *model.ProjectGeneral) string { return p.ProjectRegistrationNo }},
}

func diffGeneralFields(projectID int, prev, curr *model.ProjectGeneral, at time.Time) []model.ProjectChange {
	var changes []model.ProjectChange
	for _, f := range trackedFields {
		oldVal := f.get(prev)
		newVal := f.get(curr)
		if oldVal != newVal {
			changes = append(changes, model.ProjectChange{
				ProjectID:  projectID,
				FieldName:  f.name,
				OldValue:   oldVal,
				NewValue:   newVal,
				DetectedAt: at,
			})
		}
	}
	return changes
}

func buildProject(g *model.ProjectGeneral, promoterID *int, rawJSON []byte) *model.Project {
	p := &model.Project{
		MahaID:               g.ProjectID,
		ReraRegistrationNo:   g.ProjectRegistrationNo,
		AcknowledgementNumber: g.AcknowledgementNumber,
		ProjectName:          g.ProjectName,
		ProjectType:          g.ProjectTypeName,
		ProjectStatus:        g.ProjectStatusName,
		ProjectCurrentStatus: g.ProjectCurrentStatus,
		TotalUnits:           g.TotalNumberOfUnits,
		TotalSoldUnits:       g.TotalNumberOfSoldUnits,
		UserProfileID:        g.UserProfileID,
		PromoterID:           promoterID,
		IsMigrated:           g.IsMigrated == 1,
		RawJSON:              rawJSON,
	}

	if t := parseDate(g.ReraRegistrationDate); t != nil {
		p.ReraRegistrationDate = t
	}
	if t := parseDate(g.ProjectProposeComplitionDate); t != nil {
		p.ProposedCompletionDate = t
	}
	if t := parseDate(g.OriginalProjectProposeCompletionDate); t != nil {
		p.OriginalCompletionDate = t
	}

	return p
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return nil
	}
	return &t
}
