package lxd

import "github.com/lxc/lxd/shared/api"

func (manager *FlaskManager) createLXDStorageVolume(pool, volume, size string) error {
	createVolumeReq := api.StorageVolumesPost{
		StorageVolumePut: api.StorageVolumePut{
			Config: map[string]string{
				"size": size,
			},
		},
		Name:        volume,
		ContentType: "filesystem",
		Type:        "custom",
	}

	err := manager.client.CreateStoragePoolVolume(pool, createVolumeReq)
	if err != nil {
		manager.logger.Errorw("Unable to create LXD storage volume",
			"error", err,
		)

		return err
	}

	return nil
}

func (manager *FlaskManager) attachLXDVolumeToInstance(poolName, volumeName, instanceName, path string) error {
	instance, _, err := manager.client.GetInstance(instanceName)
	if err != nil {
		manager.logger.Errorw("Unable to get LXD instance",
			"error", err,
		)

		return err
	}

	device := map[string]string{
		"type":   "disk",
		"pool":   poolName,
		"source": volumeName,
		"path":   path,
	}

	instanceUpdateReq := instance.Writable()
	instanceUpdateReq.Devices[volumeName] = device

	op, err := manager.client.UpdateInstance(instanceName, instanceUpdateReq, "")
	if err != nil {
		manager.logger.Errorw("Unable to attach LXD storage volume to instance",
			"error", err,
		)
	}

	err = op.Wait()
	if err != nil {
		manager.logger.Errorw("An error occured while waiting for the LXD instance attach volume operation to complete",
			"error", err,
		)
	}

	return nil
}

func (manager *FlaskManager) createLXDInstance(name, image, profile string) error {
	createInstanceReq := api.InstancesPost{
		InstancePut: api.InstancePut{
			Profiles: []string{
				profile,
			},
		},
		Name: name,
		Source: api.InstanceSource{
			Type:  "image",
			Alias: image,
		},
		Type: "container",
	}

	op, err := manager.client.CreateInstance(createInstanceReq)
	if err != nil {
		manager.logger.Errorw("Unable to create LXD instance",
			"error", err,
		)

		return err
	}

	err = op.Wait()
	if err != nil {
		manager.logger.Errorw("An error occured while waiting for the LXD instance creation operation to complete",
			"error", err,
		)

		return err
	}

	return nil
}

func (manager *FlaskManager) startLXDInstance(name string) error {
	startInstanceReq := api.InstanceStatePut{
		Action:  "start",
		Timeout: 10,
	}

	op, err := manager.client.UpdateInstanceState(name, startInstanceReq, "")
	if err != nil {
		manager.logger.Errorw("Unable to start LXD instance",
			"error", err,
		)

		return err
	}

	err = op.Wait()
	if err != nil {
		manager.logger.Errorw("An error occured while waiting for the LXD instance start operation to complete",
			"error", err,
		)

		return err
	}

	return nil
}
