package hcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hetznercloud/hcloud-go/hcloud/schema"
)

func TestVolumeClientGet(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/volumes/1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"volume": {
				"id": 1,
				"created": "2016-01-30T23:50:11+00:00",
				"name": "db-storage",
				"server": null,
				"location": {
					"id": 1,
					"name": "fsn1",
					"description": "Falkenstein DC Park 1",
					"country": "DE",
					"city": "Falkenstein",
					"latitude": 50.47612,
					"longitude": 12.370071
				},
				"size": 42,
				"linux_device":"/dev/disk/by-id/scsi-0HC_volume_1",
				"protection": {
					"delete": true
				}
			}
		}`)
	})

	ctx := context.Background()

	t.Run("GetByID", func(t *testing.T) {
		volume, _, err := env.Client.Volume.GetByID(ctx, 1)
		if err != nil {
			t.Fatal(err)
		}
		if volume == nil {
			t.Fatal("no volume")
		}
		if volume.ID != 1 {
			t.Errorf("unexpected volume ID: %v", volume.ID)
		}
	})

	t.Run("Get", func(t *testing.T) {
		volume, _, err := env.Client.Volume.Get(ctx, "1")
		if err != nil {
			t.Fatal(err)
		}
		if volume == nil {
			t.Fatal("no volume")
		}
		if volume.ID != 1 {
			t.Errorf("unexpected volume ID: %v", volume.ID)
		}
	})
}

func TestVolumeClientGetByIDNotFound(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/volumes/1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(schema.ErrorResponse{
			Error: schema.Error{
				Code: string(ErrorCodeNotFound),
			},
		})
	})

	ctx := context.Background()
	volume, _, err := env.Client.Volume.GetByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if volume != nil {
		t.Fatal("expected no volume")
	}
}

func TestVolumeClientGetByName(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "name=my-volume" {
			t.Fatal("missing name query")
		}
		fmt.Fprint(w, `{
			"volumes": [
				{
					"id": 1,
					"created": "2016-01-30T23:50:11+00:00",
					"name": "my-volume",
					"server": null,
					"location": {
						"id": 1,
						"name": "fsn1",
						"description": "Falkenstein DC Park 1",
						"country": "DE",
						"city": "Falkenstein",
						"latitude": 50.47612,
						"longitude": 12.370071
					},
					"size": 42,
					"linux_device":"/dev/disk/by-id/scsi-0HC_volume_1",
					"protection": {
						"delete": true
					}
				}
			]
		}`)
	})

	ctx := context.Background()

	t.Run("GetByName", func(t *testing.T) {
		volume, _, err := env.Client.Volume.GetByName(ctx, "my-volume")
		if err != nil {
			t.Fatal(err)
		}
		if volume == nil {
			t.Fatal("no volume")
		}
		if volume.ID != 1 {
			t.Errorf("unexpected volume ID: %v", volume.ID)
		}
	})

	t.Run("Get", func(t *testing.T) {
		volume, _, err := env.Client.Volume.Get(ctx, "my-volume")
		if err != nil {
			t.Fatal(err)
		}
		if volume == nil {
			t.Fatal("no volume")
		}
		if volume.ID != 1 {
			t.Errorf("unexpected volume ID: %v", volume.ID)
		}
	})
}

func TestVolumeClientGetByNameNotFound(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "name=my-volume" {
			t.Fatal("missing name query")
		}
		fmt.Fprint(w, `{
			"volumes": []
		}`)
	})

	ctx := context.Background()
	volume, _, err := env.Client.Volume.GetByName(ctx, "my-volume")
	if err != nil {
		t.Fatal(err)
	}
	if volume != nil {
		t.Fatal("unexpected volume")
	}
}

func TestVolumeClientDelete(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/volumes/1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Error("expected DELETE")
		}
	})

	var (
		ctx    = context.Background()
		volume = &Volume{ID: 1}
	)
	_, err := env.Client.Volume.Delete(ctx, volume)
	if err != nil {
		t.Fatal(err)
	}
}

func TestVolumeClientCreateWithServer(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"volume": {
				"id": 1,
				"created": "2016-01-30T23:50:11+00:00",
				"name": "my-volume",
				"server": 1,
				"location": {
					"id": 1,
					"name": "fsn1",
					"description": "Falkenstein DC Park 1",
					"country": "DE",
					"city": "Falkenstein",
					"latitude": 50.47612,
					"longitude": 12.370071
				},
				"size": 42,
				"linux_device":"/dev/disk/by-id/scsi-0HC_volume_1",
				"protection": {
					"delete": true
				},
				"labels": {}
			},
			"action": {
				"id": 2,
				"command": "create_volume",
				"status": "running",
				"progress": 0,
				"started": "2016-01-30T23:50:11+00:00",
				"finished": null,
				"resources": [
					{
						"id": 42,
						"type": "server"
					},
					{
						"id": 1,
						"type": "volume"
					}
				]
			},
			"next_actions": [
				{
					"id": 3,
					"command": "attach_volume",
					"status": "running",
					"progress": 0,
					"started": "2016-01-30T23:50:15+00:00",
					"finished": null,
					"resources": [
						{
							"id": 42,
							"type": "server"
						},
						{
							"id": 1,
							"type": "volume"
						}
					]
				}
			]
		}`)
	})

	ctx := context.Background()
	opts := VolumeCreateOpts{
		Name:   "my-volume",
		Size:   42,
		Server: &Server{ID: 1},
	}
	result, _, err := env.Client.Volume.Create(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Volume.ID != 1 {
		t.Errorf("unexpected volume ID: %v", result.Volume.ID)
	}
	if result.Action.ID != 2 {
		t.Errorf("unexpected action ID: %v", result.Action.ID)
	}
	if len(result.NextActions) != 1 || result.NextActions[0].ID != 3 {
		t.Errorf("unexpected next actions: %v", result.NextActions)
	}
}

func TestVolumeClientCreateWithLocation(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		var reqBody schema.VolumeCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatal(err)
		}
		if reqBody.Name != "my-volume" {
			t.Errorf("unexpected volume name in request: %v", reqBody.Name)
		}
		if reqBody.Size != 42 {
			t.Errorf("unexpected volume size in request: %v", reqBody.Size)
		}
		if reqBody.Location != float64(1) {
			t.Errorf("unexpected volume location in request: %v", reqBody.Location)
		}
		if reqBody.Server != nil {
			t.Errorf("unexpected server in request: %v", reqBody.Server)
		}
		if reqBody.Labels == nil || (*reqBody.Labels)["key"] != "value" {
			t.Errorf("unexpected labels in request: %v", reqBody.Labels)
		}
		if reqBody.Automount != nil {
			t.Errorf("unexpected automount in request: %v", reqBody.Automount)
		}
		if reqBody.Format != nil {
			t.Errorf("unexpected format in request: %v", reqBody.Automount)
		}
		fmt.Fprint(w, `{
			"volume": {
				"id": 1,
				"created": "2016-01-30T23:50:11+00:00",
				"name": "my-volume",
				"server": null,
				"location": {
					"id": 1,
					"name": "fsn1",
					"description": "Falkenstein DC Park 1",
					"country": "DE",
					"city": "Falkenstein",
					"latitude": 50.47612,
					"longitude": 12.370071
				},
				"size": 42,
				"linux_device":"/dev/disk/by-id/scsi-0HC_volume_1",
				"protection": {
					"delete": true
				},
				"labels": {
					"key": "value"
				}
			}
		}`)
	})

	ctx := context.Background()
	opts := VolumeCreateOpts{
		Name:     "my-volume",
		Size:     42,
		Location: &Location{ID: 1},
		Labels:   map[string]string{"key": "value"},
	}
	result, _, err := env.Client.Volume.Create(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Volume.ID != 1 {
		t.Errorf("unexpected volume ID: %v", result.Volume.ID)
	}
}
func TestVolumeClientCreateWithAutomount(t *testing.T) {
	env := newTestEnv()
	defer env.Teardown()

	env.Mux.HandleFunc("/volumes", func(w http.ResponseWriter, r *http.Request) {
		var reqBody schema.VolumeCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatal(err)
		}
		if reqBody.Name != "my-volume" {
			t.Errorf("unexpected volume name in request: %v", reqBody.Name)
		}
		if reqBody.Size != 42 {
			t.Errorf("unexpected volume size in request: %v", reqBody.Size)
		}
		if reqBody.Server == nil || *reqBody.Server != 1 {
			t.Errorf("unexpected server in request: %v", reqBody.Server)
		}
		if reqBody.Labels == nil || (*reqBody.Labels)["key"] != "value" {
			t.Errorf("unexpected labels in request: %v", reqBody.Labels)
		}
		if *reqBody.Automount != true {
			t.Errorf("unexpected automount in request: %v", reqBody.Automount)
		}
		if *reqBody.Format != "xfs" {
			t.Errorf("unexpected format in request: %v", reqBody.Automount)
		}
		fmt.Fprint(w, `{
			"volume": {
				"id": 1,
				"created": "2016-01-30T23:50:11+00:00",
				"name": "my-volume",
				"server": 1,
				"location": {
					"id": 1,
					"name": "fsn1",
					"description": "Falkenstein DC Park 1",
					"country": "DE",
					"city": "Falkenstein",
					"latitude": 50.47612,
					"longitude": 12.370071
				},
				"size": 42,
				"linux_device":"/dev/disk/by-id/scsi-0HC_volume_1",
				"protection": {
					"delete": true
				},
				"labels": {
					"key": "value"
				}
			}
		}`)
	})

	ctx := context.Background()
	opts := VolumeCreateOpts{
		Name:      "my-volume",
		Size:      42,
		Server:    &Server{ID: 1},
		Labels:    map[string]string{"key": "value"},
		Automount: Bool(true),
		Format:    String("xfs"),
	}
	result, _, err := env.Client.Volume.Create(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	if result.Volume.ID != 1 {
		t.Errorf("unexpected volume ID: %v", result.Volume.ID)
	}
}
func TestVolumeClientUpdate(t *testing.T) {
	var (
		ctx    = context.Background()
		volume = &Volume{ID: 1}
	)

	t.Run("name", func(t *testing.T) {
		env := newTestEnv()
		defer env.Teardown()

		env.Mux.HandleFunc("/volumes/1", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Error("expected PUT")
			}
			var reqBody schema.VolumeUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Fatal(err)
			}
			if reqBody.Name != "test" {
				t.Errorf("unexpected name: %v", reqBody.Name)
			}
			json.NewEncoder(w).Encode(schema.VolumeUpdateResponse{
				Volume: schema.Volume{
					ID: 1,
				},
			})
		})

		opts := VolumeUpdateOpts{
			Name: "test",
		}
		updatedVolume, _, err := env.Client.Volume.Update(ctx, volume, opts)
		if err != nil {
			t.Fatal(err)
		}
		if updatedVolume.ID != 1 {
			t.Errorf("unexpected volume ID: %v", updatedVolume.ID)
		}
	})

	t.Run("labels", func(t *testing.T) {
		env := newTestEnv()
		defer env.Teardown()

		env.Mux.HandleFunc("/volumes/1", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Error("expected PUT")
			}
			var reqBody schema.VolumeUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Fatal(err)
			}
			if reqBody.Labels == nil || (*reqBody.Labels)["key"] != "value" {
				t.Errorf("unexpected labels in request: %v", reqBody.Labels)
			}
			json.NewEncoder(w).Encode(schema.VolumeUpdateResponse{
				Volume: schema.Volume{
					ID: 1,
				},
			})
		})

		opts := VolumeUpdateOpts{
			Name:   "test",
			Labels: map[string]string{"key": "value"},
		}
		updatedVolume, _, err := env.Client.Volume.Update(ctx, volume, opts)
		if err != nil {
			t.Fatal(err)
		}
		if updatedVolume.ID != 1 {
			t.Errorf("unexpected volume ID: %v", updatedVolume.ID)
		}
	})

	t.Run("no updates", func(t *testing.T) {
		env := newTestEnv()
		defer env.Teardown()

		env.Mux.HandleFunc("/volumes/1", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Error("expected PUT")
			}
			var reqBody schema.VolumeUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Fatal(err)
			}
			if reqBody.Name != "" {
				t.Errorf("unexpected no name, but got: %v", reqBody.Name)
			}
			if reqBody.Labels != nil {
				t.Errorf("unexpected no labels, but got: %v", reqBody.Labels)
			}
			json.NewEncoder(w).Encode(schema.VolumeUpdateResponse{
				Volume: schema.Volume{
					ID: 1,
				},
			})
		})

		opts := VolumeUpdateOpts{}
		updatedVolume, _, err := env.Client.Volume.Update(ctx, volume, opts)
		if err != nil {
			t.Fatal(err)
		}
		if updatedVolume.ID != 1 {
			t.Errorf("unexpected volume ID: %v", updatedVolume.ID)
		}
	})
}
