package lxd

import (
	"bytes"

	"github.com/gofrs/uuid"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
)

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
	id, err := uuid.NewV4()
	if err != nil {
		manager.logger.Errorw("Unable to generate UUID",
			"error", err,
		)

		return err
	}

	createInstanceReq := api.InstancesPost{
		InstancePut: api.InstancePut{
			Profiles: []string{
				profile,
			},
			Config: map[string]string{
				"user.flask-id": id.String(),
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

func (manager *FlaskManager) createFileInLXDInstance(name, path string, content []byte) error {
	createFileReq := lxd.InstanceFileArgs{
		Content:   bytes.NewReader(content),
		UID:       0,
		GID:       0,
		Mode:      0600,
		Type:      "file",
		WriteMode: "overwrite",
	}

	return manager.client.CreateInstanceFile(name, path, createFileReq)
}

func (manager *FlaskManager) getLXDInstance(name string) (*api.Instance, error) {
	instance, _, err := manager.client.GetInstance(name)
	if err != nil {
		manager.logger.Errorw("Unable to get LXD instance",
			"error", err,
		)

		return nil, err
	}

	return instance, nil
}

func (manager *FlaskManager) getLXDInstanceState(name string) (*api.InstanceState, error) {
	instanceState, _, err := manager.client.GetInstanceState(name)
	if err != nil {
		manager.logger.Errorw("Unable to get LXD instance state",
			"error", err,
		)

		return nil, err
	}

	return instanceState, nil
}

func (manager *FlaskManager) removeLXDInstance(name string) error {
	op, err := manager.client.DeleteInstance(name)
	if err != nil {
		if err != nil {
			manager.logger.Errorw("Unable to delete LXD instance",
				"error", err,
			)

			return err
		}

		err = op.Wait()
		if err != nil {
			manager.logger.Errorw("An error occured while waiting for the LXD instance delete operation to complete",
				"error", err,
			)

			return err
		}
	}

	return nil
}

func (manager *FlaskManager) removeLXDStorageVolume(pool, volume string) error {
	return manager.client.DeleteStoragePoolVolume(pool, "custom", volume)
}

func (manager *FlaskManager) startLXDInstance(name string) error {
	startInstanceReq := api.InstanceStatePut{
		Action:  "start",
		Timeout: int(manager.cfg.Client.Timeout.Seconds()),
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

func (manager *FlaskManager) stopLXDInstance(name string) error {
	stopInstanceReq := api.InstanceStatePut{
		Action:  "stop",
		Timeout: int(manager.cfg.Client.Timeout.Seconds()),
	}

	op, err := manager.client.UpdateInstanceState(name, stopInstanceReq, "")
	if err != nil {
		manager.logger.Errorw("Unable to stop LXD instance",
			"error", err,
		)

		return err
	}

	err = op.Wait()
	if err != nil {
		manager.logger.Errorw("An error occured while waiting for the LXD instance stop operation to complete",
			"error", err,
		)

		return err
	}

	return nil
}
