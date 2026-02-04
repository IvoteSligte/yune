
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

template <class... T>
struct overloaded : T... {
    using T::operator()...;
};
template <class... T>
overloaded(T...) -> overloaded<T...>;

template <class... T>
struct Union {
    // Create from element
    template <class U>
    Union(U element)
        : variant(element)
    {
    }

    // Create from subset
    template <class... U>
    Union(Union<U...> subset)
        : variant(std::visit(
              [](auto element) constexpr -> std::variant<T...> { return element.variant; },
              subset))
    {
    }

    std::variant<T...> variant;
};

using json = nlohmann::json;

inline json serialize(const std::string& s) { return s; }
inline json serialize(const int& i) { return i; }
inline json serialize(const bool& b) { return b; }
inline json serialize(const float& f) { return f; }

struct TypeType { };

struct IntType { };
struct FloatType { };
struct BoolType { };
struct StringType { };
struct TupleType;
struct ListType;
struct FnType;
struct StructType;

using namespace nlohmann::literals;

inline json serialize(const TypeType&) { return R"({ "Type": {} })"_json; }
inline json serialize(const IntType&) { return R"({ "IntType": {} })"_json; }
inline json serialize(const FloatType&) { return R"({ "FloatType": {} })"_json; }
inline json serialize(const BoolType&) { return R"({ "BoolType": {} })"_json; }
inline json serialize(const StringType&) { return R"({ "StringType": {} })"_json; }

using Type = std::variant<TypeType, IntType, FloatType, BoolType, StringType, Box<TupleType>, Box<ListType>, Box<FnType>, Box<StructType>>;

json serialize(const TypeType& t);
json serialize(const IntType& t);
json serialize(const FloatType& t);
json serialize(const BoolType& t);
json serialize(const StringType& t);
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
    explicit TupleType(std::vector<Type> elements)
    {
        this->elements = std::move(elements);
    }

    std::vector<Type> elements;
};
struct ListType {
    explicit ListType(Type element)
    {
        this->element = element;
    }

    Type element;
};
struct FnType {
    explicit FnType(Type argument, Type returnType)
    {
        this->argument = std::move(argument);
        this->returnType = std::move(returnType);
    }

    Type argument;
    Type returnType;
};
struct StructType {
    explicit StructType(std::string name)
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

template <class T>
struct Literal {
    explicit Literal(T value)
        : value(value)
    {
    }
    T value;
};
using IntegerLiteral = Literal<int>;
using FloatLiteral = Literal<float>;
using BoolLiteral = Literal<bool>;
using StringLiteral = Literal<std::string>;

inline json serialize(const IntegerLiteral& e)
{
    return { { "IntegerLiteral", { { "type", serialize(StructType("IntegerLiteral")) }, { "value", e.value } } } };
}
inline json serialize(const FloatLiteral& e)
{
    return { { "FloatLitearl", { { "type", serialize(StructType("FloatLiteral")) }, { "value", e.value } } } };
}
inline json serialize(const BoolLiteral& e)
{
    return { { "BoolLiteral", { { "type", serialize(StructType("BoolLiteral")) }, { "value", e.value } } } };
}
inline json serialize(const StringLiteral& e)
{
    return { { "StringLiteral", { { "type", serialize(StructType("StringLiteral")) }, { "value", e.value } } } };
}

template <class... T>
inline json serialize(const std::tuple<T...>& e)
{
    json elements = std::apply([](auto&&... element) { return json { serialize(element)... }; }, e);
    return { { "Tuple", { { "elements", elements } } } };
}

// TODO: other expression kinds

struct Expression {
    template <class T>
    Expression(Literal<T> value)
        : self(std::make_unique<Concrete<Literal<T>>>(std::move(value)))
    {
    }

    template <class... T>
    Expression(std::tuple<T...> value)
        : self(std::make_unique<Concrete<std::tuple<T...>>>(std::move(value)))
    {
    }

    struct Interface {
        virtual json serialize() const = 0;
        virtual ~Interface() = default;
    };
    template <class T>
    struct Concrete : Interface {
        Concrete(T value)
            : value(std::move(value))
        {
        }
        json serialize() const override
        {
            return ty::serialize(value);
        }
        T value;
    };

    template <class T>
    bool has_type() const { return dynamic_cast<T*>(self.get()) != nullptr; }
    template <class T>
    T& get() { return dynamic_cast<T*>(self.get()); }
    template <class T>
    T get() const { return *dynamic_cast<T*>(self.get()); }

    json serialize() const { return self->serialize(); }

    std::unique_ptr<Interface> self;
};

inline json serialize(const Expression& e)
{
    json j = e.serialize();
    for (auto& jStruct : j) {
        jStruct["type"] = serialize(StructType("Expression"));
    }
    return j;
}

} // namespace ty
