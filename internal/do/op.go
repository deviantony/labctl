package do

import "github.com/digitalocean/godo"

func (manager *FlaskManager) createDroplet(config dropletConfig) (int, error) {
	createRequest := &godo.DropletCreateRequest{
		Name:   config.Name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: manager.config.BaseImage,
		},
		Tags: []string{manager.config.TagName},
		SSHKeys: []godo.DropletCreateSSHKey{
			{Fingerprint: manager.config.SSHKeyFingerprint},
		},
		Monitoring: true,
	}

	newDroplet, _, err := manager.client.Droplets.Create(manager.ctx, createRequest)
	if err != nil {
		manager.logger.Errorw("Unable to create droplet",
			"error", err,
		)

		return 0, err
	}

	_, _, err = manager.client.Projects.AssignResources(manager.ctx, manager.config.ProjectID, newDroplet.URN())
	if err != nil {
		manager.logger.Warnw("Unable to assign droplet to project",
			"error", err,
		)
	}

	return newDroplet.ID, nil
}
