#pragma once

#include <format>
#include <string>
#include <iostream>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/socket.h>

constexpr int YUNE_COMPILER_PORT = 11555;

inline class CompilerConnection {
public:
  CompilerConnection() {
    socket = ::socket(AF_INET, SOCK_STREAM, 0);
    if (socket == -1) {
      std::cerr << "clang-repl: Failed to open compiler connection socket." << std::endl;
      exit(1);
    }
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = ::htons(YUNE_COMPILER_PORT);
    ::inet_pton(AF_INET, "127.0.0.1", &addr.sin_addr);

    int err = connect(socket, (sockaddr *)&addr, sizeof(addr));
    if (err != 0) {
      std::cerr << "clang-repl: Failed to connect to the compiler." << std::endl;
      exit(1);
    }
    std::cout << "clang-repl: Connected to Yune compiler." << std::endl;
  }

  ~CompilerConnection() {
    ::close(socket);
  }

  std::string get_type(std::string name) const {
    std::string payload = "getType:" + name + "\n";
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      std::cerr << "clang-repl: Failed to send a type query through the compiler connection." << std::endl;
      exit(1);
    }
    return read_line();
  }

  void yield(std::string result) const {
    std::string payload = "result:" + result + "\n";
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      std::cerr << "clang-repl: Failed to send a result through the compiler connection." << std::endl;
      exit(1);
    }
  }

private:
  std::string read_line() const {
    std::string result;
    char c;

    while (true) {
      ssize_t n = ::recv(socket, &c, 1, 0);
      if (n == -1) {
        std::cerr << "clang-repl: Failed to read line from the compiler connection." << std::endl;
        exit(1);
      }
      if (c == '\n') {
        return result;
      }
      result += c;
    }
  }

  int socket;
} compiler_connection{};
