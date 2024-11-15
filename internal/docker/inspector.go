package docker

import (
	"context"
	"encoding/json"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

type Container struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type Containers struct {
	Containers []Container `json:"containers"`
}

func GetDockerContainers() ([]byte, error) {
	log.Info("Получение статусов контейнеров в системе")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Errorf("Ошибка создания клиента docker: %v", err)
		return nil, err
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		log.Fatalf("Ошибка получения списка контейнеров: %v", err)
		return nil, err
	}

	data := Containers{}
	for _, containerObj := range containers {
		data.Containers = append(data.Containers, Container{
			Name:   containerObj.Names[0][1:],
			Status: containerObj.State,
		})
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Errorf("Ошибка формирования JSON данных о контейнера: %v", err)
		return nil, err
	}

	return jsonData, nil

}
