#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <netinet/in.h>

#define PORT 8081
#define BACKLOG 16
#define REQUEST_BUF_SIZE 4096

static const char *RESPONSE_BODY = "{\"status\":\"ok\",\"server\":\"c-raw-sockets\"}";

static void read_request(int client_fd, char *buf, size_t buf_size) {
    size_t total = 0;
    ssize_t n;

    while (total < buf_size - 1) {
        n = read(client_fd, buf + total, buf_size - 1 - total);
        if (n <= 0) {
            break;
        }
        total += (size_t)n;
        buf[total] = '\0';
        if (strstr(buf, "\r\n\r\n") != NULL) {
            break;
        }
    }
}

static void write_response(int client_fd) {
    char header[256];
    int header_len = snprintf(
        header, sizeof(header),
        "HTTP/1.1 200 OK\r\n"
        "Content-Type: application/json\r\n"
        "Content-Length: %zu\r\n"
        "Connection: close\r\n"
        "\r\n",
        strlen(RESPONSE_BODY));

    write(client_fd, header, (size_t)header_len);
    write(client_fd, RESPONSE_BODY, strlen(RESPONSE_BODY));
}

int main(void) {
    int server_fd = socket(AF_INET, SOCK_STREAM, 0);
    if (server_fd < 0) {
        perror("socket");
        return 1;
    }

    int reuse = 1;
    setsockopt(server_fd, SOL_SOCKET, SO_REUSEADDR, &reuse, sizeof(reuse));

    struct sockaddr_in addr;
    memset(&addr, 0, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_addr.s_addr = INADDR_ANY;
    addr.sin_port = htons(PORT);

    if (bind(server_fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) {
        perror("bind");
        close(server_fd);
        return 1;
    }

    if (listen(server_fd, BACKLOG) < 0) {
        perror("listen");
        close(server_fd);
        return 1;
    }

    printf("c-http-server listening on :%d\n", PORT);

    char request_buf[REQUEST_BUF_SIZE];

    for (;;) {
        struct sockaddr_in client_addr;
        socklen_t client_addr_len = sizeof(client_addr);

        int client_fd = accept(server_fd, (struct sockaddr *)&client_addr, &client_addr_len);
        if (client_fd < 0) {
            perror("accept");
            continue;
        }

        read_request(client_fd, request_buf, sizeof(request_buf));
        write_response(client_fd);

        close(client_fd);
    }

    close(server_fd);
    return 0;
}
