#pragma once

#include "pb.hpp"
#include <arpa/inet.h>
#include <iostream>
#include <semaphore>
#include <string>
#include <sys/socket.h>
#include <unistd.h>

// Alternative to std::promise<Type_t>, which gives missing symbols errors
// when used with clang-repl.
struct TypePromise_ {
  std::binary_semaphore sync{0};
  std::optional<Type_t> type;

  Type_t get() {
    this->sync.acquire();
    return std::move(*this->type);
  }

  void set(Type_t type) {
    this->type = type;
    this->sync.release();
  }
};

constexpr int YUNE_COMPILER_PORT = 11555;

inline class CompilerConnection_ {
public:
  CompilerConnection_() {
    socket = ::socket(AF_INET, SOCK_STREAM, 0);
    if (socket == -1) {
      panic("clang-repl: Failed to open compiler connection socket.");
    }
    sockaddr_in addr{};
    addr.sin_family = AF_INET;
    addr.sin_port = ::htons(YUNE_COMPILER_PORT);
    ::inet_pton(AF_INET, "127.0.0.1", &addr.sin_addr);

    int err = connect(socket, (sockaddr *)&addr, sizeof(addr));
    if (err != 0) {
      panic("clang-repl: Failed to connect to the compiler.");
    }
    std::cout << "clang-repl: Connected to Yune compiler." << std::endl;
  }

  ~CompilerConnection_() { ::close(socket); }

  Type_t get_type(String_t name) {
    std::string payload = std::format(R"({{ "getType": {} }})"
                                      "\n",
                                      toJson_(name));
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      panic("Failed to send a type query through the compiler connection.");
    }
    return type_promise.get();
  }

  void register_named_type(String_t name, Type_t type) {
    std::string payload = std::format(
        R"({{ "registerNamedType": {{ "name": {}, "value": {} }} }})"
        "\n",
        toJson_(name), toJson_(type));
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      panic("Failed to send a named type registration request through the "
            "compiler connection.");
    }
  }

  void check_named_type(String_t name) {
    std::string payload =
        std::format(R"({{ "checkNamedType": {{ "name": {} }} }})"
                    "\n",
                    toJson_(name));
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      panic("Failed to send a named type check request through the "
            "compiler connection.");
    }
  }

  void yield(std::string result) const {
    std::string payload = std::format(R"({{ "result": {} }})"
                                      "\n",
                                      result);
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      panic("Failed to send a result through the compiler connection.");
    }
  }

  void send_finished() const {
    std::string payload = R"({ "finished": {} })"
                          "\n";
    ssize_t err = ::send(socket, payload.c_str(), payload.size(), 0);
    if (err == -1) {
      panic("Failed to send a 'finished' message through the compiler "
            "connection.");
    }
  }

  void set_type(Type_t type) { type_promise.set(type); }

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
  TypePromise_ type_promise;
} compiler_connection{};

inline Type_t getType_cf(String_t name) {
  return compiler_connection.get_type(name);
}

inline void registerNamedType_cf(String_t name, Type_t type) {
  return compiler_connection.register_named_type(name, type);
}

inline void checkNamedType_cf(String_t name) {
  return compiler_connection.check_named_type(name);
}
