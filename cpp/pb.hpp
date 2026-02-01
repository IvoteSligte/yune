
#include "json.hpp" // nlohmann JSON library
#include <memory>
#include <string>
#include <vector>

namespace ty {

using json = nlohmann::json;

struct Value {
  virtual json to_json() const = 0;
};

struct Type : Value {
  virtual json to_json() const = 0;
};
using TypePtr = std::unique_ptr<Type>;

struct TypeType : Type {
  json to_json() const override { return {{"_type", "Type"}}; }
};
struct IntType : Type {
  json to_json() const override { return {{"_type", "IntType"}}; }
};
struct FloatType : Type {
  json to_json() const override { return {{"_type", "FloatType"}}; }
};
struct BoolType : Type {
  json to_json() const override { return {{"_type", "BoolType"}}; }
};
struct StringType : Type {
  json to_json() const override { return {{"_type", "StringType"}}; }
};
struct NilType : Type {
  json to_json() const override { return {{"_type", "NilType"}}; }
};
struct TupleType : Type {
  TupleType(std::vector<TypePtr> elements) {
    this->elements = std::move(elements);
  }
  json to_json() const override {
    json elementsJson;
    for (auto &ptr : elements)
      elementsJson.push_back(ptr->to_json());
    return {{"_type", "TupleType"}, {"elements", elementsJson}};
  }

  std::vector<TypePtr> elements;
};
struct ListType : Type {
  template <typename T, typename = std::enable_if_t<std::is_base_of_v<Type, T>>>
  ListType(T element) {
    this->element(std::make_unique<T>(std::move(element)));
  }
  json to_json() const override {
    return {{"_type", "ListType"}, {"element", element->to_json()}};
  }

  TypePtr element;
};
struct FnType : Type {
  template <typename A, typename B,
            typename = std::enable_if_t<std::is_base_of_v<Type, A>>,
            typename = std::enable_if_t<std::is_base_of_v<Type, B>>>
  FnType(A argument, B returnType) {
    this->argument(std::make_unique<A>(argument));
    this->returnType(std::make_unique<B>(returnType));
  }
  json to_json() const override {
    return {{"_type", "FnType"},
            {"argument", argument->to_json()},
            {"return", returnType->to_json()}};
  }

  TypePtr argument;
  TypePtr returnType;
};
struct StructType : Type {
  StructType(std::string name) : name(name) {}

  json to_json() const override {
    return {{"_type", "StructType"}, {"name", name}};
  }

  std::string name;
};

struct Expression : Value {
  virtual json to_json() const = 0;
};
struct String : Expression {
  String(std::string value) : value(value) {}

  json to_json() const override {
    return {{"_type", "String"}, {"value", value}};
  }

  std::string value;
};

inline void to_json(json &j, const Value &t) { j = t.to_json(); }
inline void to_json(json &j, const Type &t) { j = t.to_json(); }

inline std::string serializeValues(std::vector<Value> values) {
  json j = values;
  return j.dump();
}

} // namespace ty
