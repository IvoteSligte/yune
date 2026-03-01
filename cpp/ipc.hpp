#pragma once

#include "pb.hpp"
#include <string>
#include <iostream>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <future>

constexpr int YUNE_COMPILER_PORT = 11555;

static void panic(std::string message) {
  std::cerr << "clang-repl: " << message << std::endl;
  exit(1);
}

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

  ~CompilerConnection() { ::close(socket); }

  ty::Type get_type(std::string name) {
    std::string payload = std::format(R"({{ "getType": "{}" }})""\n", name);
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      panic("Failed to send a type query through the compiler connection.");
    }
    auto future = type_promise.get_future();
    future.wait(); // wait for type to be set
    return future.get();
  }
  
  void yield(std::string result) const {
    std::string payload = std::format(R"({{ "result": {} }})""\n", result);
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      panic("Failed to send a result through the compiler connection.");
    }
  }

  void set_type(ty::Type type) {
    type_promise.set_value(type);
    type_promise = {}; // reset
  }

private:
  std::string read_line() const {
    std::string result;
    char c;

    while (true) {
      ssize_t n = ::recv(socket, &c, 1, 0);
      if (n == -1) {
        panic("Failed to read line from the compiler connection.");
      }
      if (c == '\n') {
        return result;
      }
      result += c;
    }
  }

  int socket{0};
  std::promise<ty::Type> type_promise;
} compiler_connection{};

inline struct get_type_ {
  ty::Type operator()(std::string name) const {
    return compiler_connection.get_type(name);
  }
  std::string serialize() const {
    std::cerr << "get_type is not serializable as it is compile-time-only." << std::endl;
    exit(1);
  }
} get_type;

