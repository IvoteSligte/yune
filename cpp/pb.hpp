
#include "json.hpp" // nlohmann JSON library
#include <memory>
#include <string>
#include <utility>
#include <variant>
#include <vector>

template <class T>
using Box = std::shared_ptr<T>;
template <class T>
Box<T> box(T value) { return std::make_shared<T>(value); }

namespace ty {

template <class... Ts>
struct overloaded : Ts... {
    using Ts::operator()...;
};
template <class... Ts>
overloaded(Ts...) -> overloaded<Ts...>;

using json = nlohmann::json;

struct TypeType { };

struct IntType { };
struct FloatType { };
struct BoolType { };
struct StringType { };
struct NilType { };
struct TupleType;
struct ListType;
struct FnType;
struct StructType;

inline json serialize(const TypeType&) { return { { "_type", "Type" } }; }
inline json serialize(const IntType&) { return { { "_type", "IntType" } }; }
inline json serialize(const FloatType&) { return { { "_type", "FloatType" } }; }
inline json serialize(const BoolType&) { return { { "_type", "BoolType" } }; }
inline json serialize(const StringType&) { return { { "_type", "StringType" } }; }
inline json serialize(const NilType&) { return { { "_type", "NilType" } }; }

using Type = std::variant<TypeType, IntType, FloatType, BoolType, StringType, NilType, Box<TupleType>, Box<ListType>, Box<FnType>, Box<StructType>>;

struct String {
    String(std::string value)
        : value(value)
    {
    }

    std::string value;
};
inline json serialize(const String& e)
{
    return { { "_type", "String" }, { "value", e.value } };
}

using Expression = std::variant<String>;

using Value = std::variant<Type, Expression>;

json serialize(const TypeType& t);
json serialize(const IntType& t);
json serialize(const FloatType& t);
json serialize(const BoolType& t);
json serialize(const StringType& t);
json serialize(const NilType& t);
json serialize(const TupleType& t);
json serialize(const ListType& t);
json serialize(const FnType& t);
json serialize(const StructType& t);

inline json serialize(const Type& t)
{
    return std::visit(overloaded {
                          [](const Box<TupleType>& boxed) -> json { return serialize(*boxed); },
                          [](const Box<ListType>& boxed) -> json { return serialize(*boxed); },
                          [](const Box<FnType>& boxed) -> json { return serialize(*boxed); },
                          [](const Box<StructType>& boxed) -> json { return serialize(*boxed); },
                          [](const auto value) -> json { return serialize(value); },
                      },
        t);
}
inline json serialize(const Value& t)
{
    return std::visit([](auto& t) { return serialize(t); }, t);
}
inline json serialize(const Expression& t)
{
    return std::visit([](auto t) { return serialize(t); }, t);
}

struct TupleType {
    TupleType(std::vector<Type> elements)
    {
        this->elements = std::move(elements);
    }

    std::vector<Type> elements;
};
struct ListType {
    ListType(Type element)
    {
        this->element = element;
    }

    Type element;
};
struct FnType {
    FnType(Type argument, Type returnType)
    {
        this->argument = std::move(argument);
        this->returnType = std::move(returnType);
    }

    Type argument;
    Type returnType;
};
struct StructType {
    StructType(std::string name)
        : name(name)
    {
    }

    std::string name;
};

inline json serialize(const TupleType& t)
{
    json elementsJson;
    for (const Type& element : t.elements)
        elementsJson.push_back(serialize(element));
    return { { "_type", "TupleType" }, { "elements", elementsJson } };
}
inline json serialize(const ListType& t)
{
    return { { "_type", "ListType" }, { "element", serialize(t.element) } };
}
inline json serialize(const FnType& t)
{
    return {
        { "_type", "FnType" }, { "argument", serialize(t.argument) },
        { "return", serialize(t.returnType) }
    };
}
inline json serialize(const StructType& t)
{
    return { { "_type", "StructType" }, { "name", t.name } };
}

inline std::string serializeValues(std::vector<Value> values)
{
    json j;
    for (auto& v : values)
        j.push_back(serialize(v));
    return j.dump();
}

} // namespace ty
