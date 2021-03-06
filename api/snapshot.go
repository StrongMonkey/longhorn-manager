package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rancher/go-rancher/api"
)

func (s *Server) SnapshotCreate(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		err = errors.Wrap(err, "fail to create snapshot")
	}()
	var input SnapshotInput

	apiContext := api.GetApiContext(req)
	if err := apiContext.Read(&input); err != nil {
		return err
	}

	for k, v := range input.Labels {
		if strings.Contains(k, "=") || strings.Contains(v, "=") {
			return fmt.Errorf("labels cannot contain '='")
		}
	}

	volName := mux.Vars(req)["name"]
	if volName == "" {
		return fmt.Errorf("volume name required")
	}

	engine, err := s.m.GetEngineClient(volName)
	if err != nil {
		return err
	}
	snapName, err := engine.SnapshotCreate(input.Name, input.Labels)
	if err != nil {
		return err
	}

	snap, err := engine.SnapshotGet(snapName)
	if err != nil {
		return err
	}
	if snap == nil {
		return fmt.Errorf("cannot found just created snapshot '%s', for volume '%s'", snapName, volName)
	}
	apiContext.Write(toSnapshotResource(snap))
	return nil
}

func (s *Server) SnapshotList(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		err = errors.Wrap(err, "fail to list snapshot")
	}()

	volName := mux.Vars(req)["name"]
	if volName == "" {
		return fmt.Errorf("volume name required")
	}

	engine, err := s.m.GetEngineClient(volName)
	if err != nil {
		return err
	}
	snapList, err := engine.SnapshotList()
	if err != nil {
		return err
	}
	api.GetApiContext(req).Write(toSnapshotCollection(snapList))
	return nil
}

func (s *Server) SnapshotGet(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		err = errors.Wrap(err, "fail to get snapshot")
	}()

	var input SnapshotInput

	apiContext := api.GetApiContext(req)
	if err := apiContext.Read(&input); err != nil {
		return err
	}
	if input.Name == "" {
		return fmt.Errorf("empty snapshot name not allowed")
	}

	volName := mux.Vars(req)["name"]
	if volName == "" {
		return fmt.Errorf("volume name required")
	}

	engine, err := s.m.GetEngineClient(volName)
	if err != nil {
		return err
	}
	snap, err := engine.SnapshotGet(input.Name)
	if err != nil {
		return err
	}
	if snap == nil {
		return fmt.Errorf("cannot find snapshot '%s' for volume '%s'", input.Name, volName)
	}
	api.GetApiContext(req).Write(toSnapshotResource(snap))
	return nil
}

func (s *Server) SnapshotDelete(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		err = errors.Wrap(err, "fail to delete snapshot")
	}()

	var input SnapshotInput

	apiContext := api.GetApiContext(req)
	if err := apiContext.Read(&input); err != nil {
		return err
	}
	if input.Name == "" {
		return fmt.Errorf("empty snapshot name not allowed")
	}

	volName := mux.Vars(req)["name"]
	if volName == "" {
		return fmt.Errorf("volume name required")
	}

	engine, err := s.m.GetEngineClient(volName)
	if err != nil {
		return err
	}
	if err := engine.SnapshotDelete(input.Name); err != nil {
		return err
	}

	snap, err := engine.SnapshotGet(input.Name)
	if err != nil {
		return err
	}
	if snap == nil {
		return fmt.Errorf("cannot find snapshot '%s', for volume '%s'", input.Name, volName)
	}

	api.GetApiContext(req).Write(toSnapshotResource(snap))
	return nil
}

func (s *Server) SnapshotRevert(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		err = errors.Wrap(err, "fail to revert snapshot")
	}()

	var input SnapshotInput

	apiContext := api.GetApiContext(req)
	if err := apiContext.Read(&input); err != nil {
		return err
	}
	if input.Name == "" {
		return fmt.Errorf("empty snapshot name not allowed")
	}

	volName := mux.Vars(req)["name"]
	if volName == "" {
		return fmt.Errorf("volume name required")
	}

	engine, err := s.m.GetEngineClient(volName)
	if err != nil {
		return err
	}
	if err := engine.SnapshotRevert(input.Name); err != nil {
		return err
	}

	snap, err := engine.SnapshotGet(input.Name)
	if err != nil {
		return err
	}
	if snap == nil {
		return fmt.Errorf("not found snapshot '%s', for volume '%s'", input.Name, volName)
	}

	api.GetApiContext(req).Write(toSnapshotResource(snap))
	return nil
}

func (s *Server) SnapshotBackup(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		err = errors.Wrap(err, "fail to backup snapshot")
	}()

	var input SnapshotInput

	apiContext := api.GetApiContext(req)
	if err := apiContext.Read(&input); err != nil {
		return err
	}
	if input.Name == "" {
		return fmt.Errorf("empty snapshot name not allowed")
	}

	volName := mux.Vars(req)["name"]
	if volName == "" {
		return fmt.Errorf("volume name required")
	}

	settings, err := s.m.SettingsGet()
	if err != nil || settings == nil {
		return fmt.Errorf("cannot backup: unable to read settings")
	}
	backupTarget := settings.BackupTarget
	if backupTarget == "" {
		return fmt.Errorf("cannot backup: backupTarget not set")
	}

	if err := s.m.SnapshotBackup(volName, input.Name, backupTarget); err != nil {
		return err
	}

	return nil
}

func (s *Server) SnapshotPurge(w http.ResponseWriter, req *http.Request) (err error) {
	defer func() {
		err = errors.Wrap(err, "fail to purge snapshot")
	}()

	volName := mux.Vars(req)["name"]
	if volName == "" {
		return fmt.Errorf("volume name required")
	}

	if err := s.m.SnapshotPurge(volName); err != nil {
		return err
	}
	return nil
}
