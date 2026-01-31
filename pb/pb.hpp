#include <vector>
#include <string>
#include <alpaca/alpaca.h>
#include <iostream>

struct Value {
  bool operator==(const Value&) const = default;
};

struct Type : Value {};
struct TypeType : Type {};
struct IntType : Type {};
struct FloatType : Type {};
struct BoolType : Type {};
struct StringType : Type {};
struct NilType : Type {};
struct TupleType : Type {
  TupleType(std::vector<Type> elements) : elements(elements) {}
  
  std::vector<Type> elements;
};
struct ListType : Type {
  ListType(Type element) : element(element) {}
  
  Type element;
};
struct FnType : Type {
  FnType(Type argument, Type returnType) : argument(argument), returnType(returnType) {}
  
  Type argument;
  Type returnType;
};
struct StructType : Type {
  StructType(std::string name) : name(name) {}
  
  std::string name;
};

struct Expression : Value {};
struct String : Expression {
  std::string value;
};

std::vector<char> serializeValue(Value value) {
  std::vector<char> bytes;
  alpaca::serialize(value, bytes);
  return bytes;
}

Value deserializeValue(std::vector<char> bytes) {
  std::error_code error_code;
  auto value = alpaca::deserialize<Value>(bytes, error_code);
  if (error_code) {
    std::cerr << "Error deserializing Value: " << error_code << std::endl;
    exit(1);
  }
  return value;
}

