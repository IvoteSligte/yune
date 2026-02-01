
#include "json.hpp" // nlohmann JSON library
#include <memory>
#include <string>
#include <utility>
#include <variant>
#include <vector>

template <class T>
using Box = std::unique_ptr<T>;
template <class T>
Box<T> box(T value) { return std::make_unique<T>(value); }

namespace ty {

using json = nlohmann::json;

struct TypeType;
struct IntType;
struct FloatType;
struct BoolType;
struct StringType;
struct NilType;
struct TupleType;
struct ListType;
struct FnType;
struct StructType;

using Type = std::variant<TypeType, IntType, FloatType, BoolType, StringType,
    NilType, TupleType, ListType, FnType, StructType>;

struct String;
using Expression = std::variant<String>;

using Value = std::variant<Type, Expression>;

using TypePtr = Box<Type>;

struct TypeType {
    json to_json() const { return { { "_type", "Type" } }; }
};
struct IntType {
    json to_json() const { return { { "_type", "IntType" } }; }
};
struct FloatType {
    json to_json() const { return { { "_type", "FloatType" } }; }
};
struct BoolType {
    json to_json() const { return { { "_type", "BoolType" } }; }
};
struct StringType {
    json to_json() const { return { { "_type", "StringType" } }; }
};
struct NilType {
    json to_json() const { return { { "_type", "NilType" } }; }
};
struct TupleType {
    TupleType(std::vector<Type> elements)
    {
        this->elements = std::move(elements);
    }
    json to_json() const
    {
        json elementsJson;
        for (auto& t : elements)
          elementsJson.push_back(std::visit([](auto t) {
                return t.to_json(); }, t);
        return { { "_type", "TupleType" }, { "elements", elementsJson } };
    }

    std::vector<Type> elements;
};
struct ListType {
    template <typename T, typename = std::enable_if_t<std::is_base_of_v<Type, T>>>
    ListType(T element)
    {
        this->element(box<T>(std::move(element)));
    }
    json to_json() const
    {
        return { { "_type", "ListType" }, { "element", element->to_json() } };
    }

    Type element;
};
struct FnType {
    FnType(Type argument, Type returnType)
    {
        this->argument(argument);
        this->returnType(returnType);
    }
    json to_json() const
    {
        return { { "_type", "FnType" },
            { "argument", std::visit([](auto& arg) -> json { return arg.to_json(); }, *argument) },
            { "return", returnType->to_json() } };
    }

    Type argument;
    Type returnType;
};
struct StructType {
    StructType(std::string name)
        : name(name)
    {
    }

    json to_json() const { return { { "_type", "StructType" }, { "name", name } }; }

    std::string name;
};

struct String {
    String(std::string value)
        : value(value)
    {
    }

    json to_json() const { return { { "_type", "String" }, { "value", value } }; }

    std::string value;
};

inline void to_json(json& j, const Type& t)
{
    j = std::visit([](auto t) { return t.to_json(); }, t);
}
inline void to_json(json& j, const Expression& t)
{
    j = std::visit([](auto t) { return t.to_json(); }, t);
}
inline void to_json(json& j, const Value& t)
{
    j = std::visit([](auto t) { return to_json(t); }, t);
}

inline std::string serializeValues(std::vector<Value> values)
{
    json j = values;
    p return j.dump();
}

} // namespace ty
