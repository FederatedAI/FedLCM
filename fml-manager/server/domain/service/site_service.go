// Copyright 2022 VMware, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"time"

	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/entity"
	"github.com/FederatedAI/FedLCM/fml-manager/server/domain/repo"
	"github.com/FederatedAI/FedLCM/fml-manager/server/infrastructure/event"
	"github.com/FederatedAI/FedLCM/fml-manager/server/infrastructure/siteportal"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type SiteService struct {
	// SiteRepo is the repository for persisting site info
	SiteRepo repo.SiteRepository
}

// HandleSiteRegistration creates or updates the site info
func (s *SiteService) HandleSiteRegistration(site *entity.Site) error {
	if site.UUID == "" {
		return errors.New("invalid uuid")
	}
	if site.ExternalHost == "" || site.ExternalPort == 0 {
		return errors.New("invalid site connection info")
	}
	if site.Name == "" {
		return errors.New("invalid site name")
	}
	// check the connection with site
	client := siteportal.NewSitePortalClient(site.ExternalHost, site.ExternalPort, site.HTTPS, site.ServerName)
	err := client.CheckSiteStatus()
	if err != nil {
		return errors.Wrapf(err, "fml manager can not connect to site")
	}
	// reset the gorm.Model fields
	site.Model = gorm.Model{}
	site.LastRegisteredAt = time.Now()
	exist, err := s.SiteRepo.ExistByUUID(site.UUID)
	if err != nil {
		return errors.Wrap(err, "failed to find site info")
	}
	if exist {
		log.Info().Msgf("deleting stale site info with uuid: %s", site.UUID)
		if err := s.SiteRepo.DeleteByUUID(site.UUID); err != nil {
			return err
		}
	}
	log.Info().Msgf("creating site: %s with uuid: %s", site.Name, site.UUID)
	_, err = s.SiteRepo.Save(site)

	// send the site info update event to the project context
	go func() {
		log.Info().Msgf("sending site info update event to project context: site %s(%s)", site.Name, site.UUID)
		if err := event.NewSelfHttpExchange().PostEvent(event.ProjectParticipantUpdateEvent{
			UUID:        site.UUID,
			PartyID:     site.PartyID,
			Name:        site.Name,
			Description: site.Description,
		}); err != nil {
			log.Err(err).Msgf("failed to post site info update event")
		}
	}()
	return err
}
