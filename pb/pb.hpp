#include <vector>
#include <string>
#include "rfl/json.hpp"
#include "rfl.hpp"
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

std::string serializeValues(std::vector<Value> values) {
  return rfl::json::write(values);
}

std::vector<Value> deserializeValues(std::string data) {
  return rfl::json::read<std::vector<Value>>(data).value();
}

