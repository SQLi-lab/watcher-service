package docker

import (
	"context"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

func GetDockerContainers() (map[string]string, error) {
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

	data := make(map[string]string)
	for _, containerObj := range containers {
		data[containerObj.Names[0][1:]] = containerObj.State
	}

	return data, nil

}
