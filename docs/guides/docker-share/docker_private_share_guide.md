---
sidebar_position: 20
sidebar_label: Private Share
---


# Docker Private Share

With zrok, you can privately share a server app that's running in Docker, or any server that's reachable by the zrok container. Then, a zrok private access running somewhere else can use the private share. In this guide we'll cover both sides: the private share and the private access.

## Walkthrough Video

<iframe width="100%" height="315" src="https://www.youtube.com/embed/HxyvtFAvwUE" title="YouTube video player" frameborder="0" allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share" allowfullscreen></iframe>

## Before You Begin

To follow this guide you will need [Docker](https://docs.docker.com/get-docker/) and [the Docker Compose plugin](https://docs.docker.com/compose/install/) for running `docker compose` commands in your terminal.

If you have installed Docker Desktop on macOS or Windows then you are all set.

## Private Share with Docker Compose

First, let's create the private share.

1. Make a folder on your computer to use as a Docker Compose project for your zrok private share.
1. In your terminal, change directory to your newly-created project folder.
1. Download [the zrok-private-share Docker Compose project file](pathname:///zrok-private-share/compose.yml) into your new project folder and make sure it's named `compose.yml`.
1. Copy your zrok environment token from the zrok web console to your clipboard and paste it in a file named `.env` in the same folder like this:

    ```bash
    # file name ".env"
    ZROK_ENABLE_TOKEN="8UL9-48rN0ua"
    ```

1. If you are self-hosting zrok then it's important to set your API endpoint URL too. If you're using the hosted zrok service then you can skip this step.

    ```bash
    # file name ".env"
    ZROK_API_ENDPOINT="https://zrok.example.com"
    ```

1. Run your Compose project to start sharing the built-in demo web server:

    ```bash
    docker compose up
    ```

1. Read the private share token from the output. One of the last lines is like this:

    ```bash
    zrok-private-share-1  | zrok access private wr3hpf2z5fiy
    ```

    Keep track of this token so you can use it in your zrok private access project.

## Private Access with Docker Compose

Now that we have a private share we can access it with zrok running in Docker. Next, let's access the demo web server in a web browser.

1. Make a folder on your computer to use as a Docker Compose project for your zrok private access.
1. In your terminal, change directory to your newly-created project folder.
1. Download [the zrok-private-access Docker Compose project file](pathname:///zrok-private-access/compose.yml) into your new project folder and make sure it's named `compose.yml`.
1. Copy your zrok environment token from the zrok web console to your clipboard and paste it in a file named `.env` in the same folder like this:

    ```bash
    # file name ".env"
    ZROK_ENABLE_TOKEN="8UL9-48rN0ua"
    ```

1. Now copy the zrok private access token from the zrok private share project's output to your clipboard and paste it in the same file named `.env` here in your private share project folder like this:

    ```bash
    # file name ".env"
    ZROK_ENABLE_TOKEN="8UL9-48rN0ua"
    ZROK_ACCESS_TOKEN="wr3hpf2z5fiy"
    ```

1. Run your Compose project to start accessing the private share:

    ```bash
    docker compose up zrok-private-access
    ```

1. Now your zrok private access proxy is ready on http://127.0.0.1:9191. You can visit the demo web server in your browser.

## Going Further with Private Access

1. Try changing the demo web server used in the private share project. One alternative demo server is provided: `httpbin`.
1. Try accessing the private share from _inside_ a container running in the private access project. One demo client is provided: `demo-client`. You can run it like this.

    ```bash
    docker compose up demo-client
    ```

1. You'll see in the terminal output that the demo-client container is getting a response from the private share indicating the source IP of the request from the perspective of the demo server: `httpbin` that's running in the private share project.

## Cleaning Up

Run the "down" command in both Compose projects to destroy them when you're all done. This will stop the running containers and delete zrok environments' storage volumes. Then delete the selected zrok environment by clicking "Actions" in the web console.

```bash
docker compose down --remove-orphans --volumes
```
