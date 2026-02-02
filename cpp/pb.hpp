
#include "json.hpp" // nlohmann JSON library
#include <memory>
#include <string>
#include <tuple>
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

using namespace nlohmann::literals;

inline json serialize(const TypeType&) { return R"({ "Type": {} })"_json; }
inline json serialize(const IntType&) { return R"({ "IntType": {} })"_json; }
inline json serialize(const FloatType&) { return R"({ "FloatType": {} })"_json; }
inline json serialize(const BoolType&) { return R"({ "BoolType", {} })"_json; }
inline json serialize(const StringType&) { return R"({ "StringType": {} })"_json; }
inline json serialize(const NilType&) { return R"({ "NilType", {} })"_json; }

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
    return { { "String", { { "value", e.value } } } };
}

struct Expression {
    Expression(String value)
        : value(value)
    {
    }

    template <class T>
    bool has_type() const { return value.type() == typeid(T); }
    template <class T>
    T& get() { return std::any_cast<T>(value); }
    template <class T>
    T get() const { return std::any_cast<T>(value); }

    std::any value;
};

inline json serialize(const Expression& e)
{
    if (e.has_type<String>()) {
        return ty::serialize(e.get<String>());
    }
    exit(1); // unreachable
}

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
    return { { "TupleType", { { "elements", elementsJson } } } };
}
inline json serialize(const ListType& t)
{
    return { { "ListType", { { "element", serialize(t.element) } } } };
}
inline json serialize(const FnType& t)
{
    return { { "FnType", { { "argument", serialize(t.argument) }, { "return", serialize(t.returnType) } } } };
}
inline json serialize(const StructType& t)
{
    return { { "StructType", { { "name", t.name } } } };
}

template <class... T>
inline json serialize(const std::tuple<T...>& t)
{
    return serialize(Tuple((std::get<T>(t), ...)));
}

} // namespace ty
