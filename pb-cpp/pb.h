#include <vector>
#include <string>
#include <alpaca/alpaca.h>
#include <iostream>

struct Value {
  std::vector<char> serialize() {
    std::vector<char> bytes;
    alpaca::serialize(*this, bytes);
    return bytes;
  }

  static Value deserialize(std::vector<char> bytes) {
    std::error_code error_code;
    auto value = alpaca::deserialize<Value>(bytes, error_code);
    if (error_code) {
      std::cerr << "Error deserializing Value: " << error_code << std::endl;
      exit(1);
    }
    return value;
  }
};

struct Type : Value {};
struct IntType : Type {};
struct Float : Type {};
struct BoolType : Type {};
struct StringType : Type {};
struct TupleType : Type {
  std::vector<Type> elements;
};
struct ListType : Type {
  Type element;
};
struct FnType : Type {
  Type argument;
  Type _return;
};
struct StructType : Type {
  std::string name;
};

struct Expression : Value {};
struct String : Expression {
  std::string value;
};
