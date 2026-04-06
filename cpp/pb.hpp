#pragma once

// headers also used by Yune programs
#include <algorithm>
#include <iostream> // std::cout
#include <set>
#include <string> // std::string
#include <tuple>  // std::tuple, std::apply
#include <vector> // std::vector

// headers for this file
#include <concepts>
#include <format>
#include <iomanip>
#include <memory>
#include <sstream>
#include <string>
#include <utility>
#include <variant>

namespace ty {
template <class T> struct Box {
  constexpr Box(T *ptr) : ptr(ptr) {}

  Box(T &&value)
      : ptr(std::make_shared<std::decay_t<T>>(std::forward<T>(value))) {}

  Box(std::shared_ptr<T> &&ptr) : ptr(ptr) {}

  bool operator==(const Box<T> &other) const {
    return this->get() == other.get(); // compare inner values (not pointers)
  }

  T &get() const {
    if (std::holds_alternative<T *>(ptr)) {
      return *std::get<T *>(ptr);
    } else {
      return *std::get<std::shared_ptr<T>>(ptr);
    }
  }

  std::variant<std::shared_ptr<T>, T *> ptr;
};

template <class T> Box<T> box(T *value) { return Box(value); }

template <class T> Box<T> box(T &&value) {
  return std::make_shared<std::decay_t<T>>(std::forward<T>(value));
}

// Checks if T is among the classes Ts
template <class T, class... Ts>
inline constexpr bool is_one_of_v = (std::is_same_v<T, Ts> || ...);

template <class... T> struct Union {
  // Create from element directly
  template <class U>
    requires is_one_of_v<std::decay_t<U>, T...>
  constexpr Union(U &&element) : variant(std::forward<U>(element)) {}

  // Create from subset
  template <class... U>
  constexpr Union(const Union<U...> &subset)
      : variant(std::visit([](auto &&element) constexpr
                               -> std::variant<T...> { return element; },
                           subset.variant)) {}

  bool operator==(const Union<T...> &other) const = default;

  std::variant<T...> variant;
};

// Specialization of Union for zero elements.
// This is not constructable in Yune, but required for certain type signatures.
template <> struct Union<> {
  bool operator==(const Union<> &other) const = default;
};

template <class T> struct List {
  struct ArrayRef {
    size_t size;
    T *ptr;
  };

  constexpr List() = default;

  template <size_t N>
  constexpr List(T array[N]) : value(ArrayRef{.size = N, .ptr = array}) {}

  List(std::initializer_list<T> list) : value(std::vector(list)) {}

  List(std::vector<T> value) : value(value) {}

  size_t size() const {
    if (std::holds_alternative<std::vector<T>>(value)) {
      return std::get<std::vector<T>>(value).size();
    } else {
      return std::get<ArrayRef>(value).size;
    }
  }

  // Returns a new list, which is the copied contents of this list with the new
  // element appended.
  // Note that there is no `push_back` function as in `std::vector` because this
  // list may not be owned.
  List<T> append(T element) const {
    std::vector<T> result;

    if (std::holds_alternative<std::vector<T>>(value)) {
      result = std::get<std::vector<T>>(value);
    } else {
      auto array = std::get<ArrayRef>(value);
      result = std::vector<T>(array.ptr, array.ptr + array.size);
    }
    result.push_back(element);
    return result;
  }

  T *begin() {
    if (std::holds_alternative<std::vector<T>>(value)) {
      return std::get<std::vector<T>>(value).data();
    } else {
      return std::get<ArrayRef>(value).ptr;
    }
  }

  T *end() {
    if (std::holds_alternative<std::vector<T>>(value)) {
      auto &vector = std::get<std::vector<T>>(value);
      return vector.data() + vector.size();
    } else {
      auto &array = std::get<ArrayRef>(value);
      return array.ptr + array.size;
    }
  }

  const T &at(size_t index) const {
    if (std::holds_alternative<std::vector<T>>(value)) {
      return std::get<std::vector<T>>(value)[index];
    } else {
      auto array = std::get<ArrayRef>(value);
      if (index < 0 || index >= array.size) {
        std::cerr << "Out-of-bounds ArrayRef access." << std::endl;
        abort();
      }
      return std::get<ArrayRef>(value).ptr[index];
    }
  }

  const T &operator[](size_t index) const { return at(index); }

  bool operator==(const List<T> &other) const {
    if (other.size() != size()) {
      return false;
    }
    for (size_t i = 0; i < size(); i++) {
      if (at(i) != other.at(i)) {
        return false;
      }
    }
    return true;
  }

  std::variant<std::vector<T>, ArrayRef> value;
};

// Immutable string datatype.
struct String {
  constexpr String(const char *string) : value(string) {}
  constexpr String(std::string string) : value(string) {}

  std::string to_owned() const {
    return std::string(begin(), end());
  }
  
  size_t length() const {
    if (std::holds_alternative<std::string>(value)) {
      return std::get<std::string>(value).length();
    } else {
      return std::string(std::get<const char *>(value)).length();
    }
  }

  String subString(int start, int end) const {
    if (std::holds_alternative<std::string>(value)) {
      return std::get<std::string>(value).substr(start, end - start);
    } else {
      return std::string(std::get<const char *>(value), start, end - start);
    }
  }

  String operator+(const String &other) const {
    std::string concat;

    if (std::holds_alternative<std::string>(value)) {
      concat = std::get<std::string>(value);
    } else {
      concat = std::string(std::get<const char *>(value));
    }
    if (std::holds_alternative<std::string>(other.value)) {
      concat += std::get<std::string>(other.value);
    } else {
      concat += std::get<const char *>(other.value);
    }
    return concat;
  }

  bool operator==(const String &other) const {
    std::string left;
    if (std::holds_alternative<std::string>(value)) {
      left = std::get<std::string>(value);
    } else {
      left = std::string(std::get<const char *>(value));
    }
    if (std::holds_alternative<std::string>(other.value)) {
      return left == std::get<std::string>(other.value);
    } else {
      return left == std::string(std::get<const char *>(other.value));
    }
  }

  const char *begin() const {
    if (std::holds_alternative<std::string>(value)) {
      return std::get<std::string>(value).data();
    } else {
      return std::get<const char *>(value);
    }
  }

  const char *end() const { return begin() + length(); }

  // Either a C++ allocated std::string or non-owned C-string.
  std::variant<std::string, const char *> value;
};

inline std::ostream &operator<<(std::ostream &os, const ty::String &s) {
  for (const char c : s) {
    os << c;
  }
  return os;
}

template <class F, class Return, class... Args>
concept FunctionLike = requires(F f, Args... args) {
  { f(std::forward<Args>(args)...) } -> std::convertible_to<Return>;
  { f.serialize() } -> std::convertible_to<std::string>;
};

// FIXME: this is probably not runtime-free
// A serializable function class similar to std::function
template <class Return, class... Args> struct Function {
  struct Concept {
    virtual ~Concept() = default;
    virtual Return operator()(Args &&...args) const = 0;
    virtual std::string serialize() const = 0;
  };
  template <class F> struct Model final : Concept {
    explicit Model(F f) : function(std::move(f)) {}

    Return operator()(Args &&...args) const override {
      return function(std::forward<Args>(args)...);
    }
    std::string serialize() const override { return function.serialize(); }

    F function;
  };
  template <class F>
    requires FunctionLike<F, Return, Args...>
  Function(F function)
      : self(std::make_shared<std::decay_t<Model<F>>>(std::move(function))) {
    static_assert(std::is_class_v<F>,
                  "Function requires callable object, not function pointer");
  }

  Return operator()(Args... args) const {
    return (*self)(std::forward<Args>(args)...);
  }
  std::string serialize() const { return self->serialize(); }
  // TODO: copy and move operators

  std::shared_ptr<Concept> self;
};

// extends std::apply to work for a zero-sized tuple
template <class F, class Tuple> decltype(auto) apply(F &&f, Tuple &&tuple) {
  if constexpr (std::tuple_size_v<std::remove_reference_t<Tuple>> == 0) {
    return std::forward<F>(f)();
  } else {
    return std::apply(std::forward<F>(f), std::forward<Tuple>(tuple));
  }
}

struct Span {
  Span(int line, int column) : line(line), column(column) {}
  int line;
  int column;
};

struct TypeType {
  bool operator==(const TypeType &other) const { return true; }
};
struct IntType {
  bool operator==(const IntType &other) const { return true; }
};
struct FloatType {
  bool operator==(const FloatType &other) const { return true; }
};
struct BoolType {
  bool operator==(const BoolType &other) const { return true; }
};
struct StringType {
  bool operator==(const StringType &other) const { return true; }
};
struct TupleType;
struct ListType;
struct FnType;
struct StructType;
struct UnionType;

using Type =
    Union<TypeType, IntType, FloatType, BoolType, StringType, Box<TupleType>,
          Box<ListType>, Box<FnType>, Box<StructType>, Box<UnionType>>;

struct TupleType {
  ty::List<Type> elements;
  bool operator==(const TupleType &other) const = default;
};
struct ListType {
  Type element;
  bool operator==(const ListType &other) const = default;
};
struct FnType {
  Type argument;
  Type returnType;
  bool operator==(const FnType &other) const = default;
};
struct StructType {
  String name;
  bool operator==(const StructType &other) const = default;
};
struct UnionType {
  ty::List<Type> variants;
  bool operator==(const UnionType &other) const = default;
};

template <class T> struct Literal {
  T value;
};
using IntegerLiteral = Literal<int>;
using FloatLiteral = Literal<float>;
using BoolLiteral = Literal<bool>;
using StringLiteral = Literal<String>;
struct Variable;
struct FunctionCall;
struct TupleExpression;
struct UnaryExpression;
struct BinaryExpression;

using Expression =
    Union<IntegerLiteral, FloatLiteral, BoolLiteral, StringLiteral, Variable,
          Box<FunctionCall>, Box<TupleExpression>, Box<UnaryExpression>,
          Box<BinaryExpression>>;

struct Variable {
  String name;
};
struct FunctionCall {
  Expression function;
  Expression argument;
};
struct TupleExpression {
  ty::List<Expression> elements;
};
struct UnaryExpression {
  String op;
  Expression expression;
};
struct BinaryExpression {
  String op;
  Expression left;
  Expression right;
};

struct VariableDeclaration;
struct Assignment;
struct BranchStatement;

using Statement =
    Union<Box<VariableDeclaration>, Box<Assignment>, Box<BranchStatement>>;

using Block = ty::List<Statement>;

struct VariableDeclaration {
  String name;
  Block body;
};
struct Assignment {
  Variable target;
  String op;
  Block body;
};
struct BranchStatement {
  Expression condition;
  Block thenBlock;
  Block elseBlock;
};

// Escape string to JSON literal.
inline std::string serialize(const String &s) {
  std::ostringstream oss;
  oss << '"';

  for (const char c : s) {
    switch (c) {
    case '"':
      oss << "\\\"";
      break;
    case '\\':
      oss << "\\\\";
      break;
    case '\b':
      oss << "\\b";
      break;
    case '\f':
      oss << "\\f";
      break;
    case '\n':
      oss << "\\n";
      break;
    case '\r':
      oss << "\\r";
      break;
    case '\t':
      oss << "\\t";
      break;
    default:
      if (c < 0x20) {
        oss << "\\u" << std::hex << std::setw(4) << std::setfill('0')
            << static_cast<int>(c);
      } else {
        oss << c;
      }
    }
  }
  oss << '"';
  return oss.str();
}
inline std::string serialize(const int &i) { return std::to_string(i); }
inline std::string serialize(const bool &b) { return std::to_string(b); }
inline std::string serialize(const float &f) { return std::to_string(f); }

inline std::string serialize(const TypeType &) {
  return R"({ "TypeType": {} })";
};
inline std::string serialize(const IntType &) { return R"({ "IntType": {} })"; }
inline std::string serialize(const FloatType &) {
  return R"({ "FloatType": {} })";
}
inline std::string serialize(const BoolType &) {
  return R"({ "BoolType": {} })";
}
inline std::string serialize(const StringType &) {
  return R"({ "StringType": {} })";
}
std::string serialize(const TupleType &t);
std::string serialize(const ListType &t);
std::string serialize(const FnType &t);
std::string serialize(const StructType &t);
std::string serialize(const UnionType &t);
template <class T> std::string serialize(ty::List<T> elements);
template <class... T> std::string serialize(std::tuple<T...> elements);
template <class... T> std::string serialize(const Union<T...> &u);
std::string serialize(const IntegerLiteral &e);
std::string serialize(const FloatLiteral &e);
std::string serialize(const BoolLiteral &e);
std::string serialize(const StringLiteral &e);
std::string serialize(const Variable &e);
std::string serialize(const FunctionCall &e);
std::string serialize(const TupleExpression &e);
std::string serialize(const UnaryExpression &e);
std::string serialize(const BinaryExpression &e);
std::string serialize(const VariableDeclaration &e);
std::string serialize(const Assignment &e);
std::string serialize(const BranchStatement &e);
// Fallback for classes that have a serialize() method.
template <class T> std::string serialize(T object);

template <class T> std::string serialize(List<T> elements) {
  std::ostringstream oss;
  oss << '[';
  for (int i = 0; i < elements.size(); i++) {
    oss << serialize(elements[i]);
    if (i + 1 < elements.size()) {
      oss << ", ";
    }
  }
  oss << ']';
  return oss.str();
}

template <class... T> std::string serialize(std::tuple<T...> tuple) {
  std::ostringstream oss;
  oss << R"({ "Tuple": { "elements": [)";
  int i = 0;
  std::apply(
      [&](auto &&...elements) {
        (([&]() {
           oss << ty::serialize(elements);
           if (i + 1 < sizeof...(T)) {
             oss << ", ";
           }
           i++;
         }()),
         ...);
      },
      tuple);
  oss << "] } }";
  return oss.str();
}

inline std::string serialize(const TupleType &t) {
  return R"({ "TupleType": { "elements": )" + ty::serialize(t.elements) +
         " } }";
}
inline std::string serialize(const ListType &t) {
  return R"({ "ListType": { "element": )" + ty::serialize(t.element) + " } }";
}
inline std::string serialize(const FnType &t) {
  return R"({ "FnType": { "argument": )" + ty::serialize(t.argument) +
         R"(, "return": )" + ty::serialize(t.returnType) + " } }";
}
inline std::string serialize(const StructType &t) {
  return R"({ "StructType": { "name": )" + ty::serialize(t.name) + " } }";
}
inline std::string serialize(const UnionType &t) {
  return R"({ "UnionType": { "variants": )" + ty::serialize(t.variants) +
         " } }";
}

template <class T>
inline std::string serialize(const Literal<T> &literal, std::string name) {
  return R"({ ")" + name + R"(": { "value": )" + ty::serialize(literal.value) +
         " } }";
}
inline std::string serialize(const IntegerLiteral &e) {
  return ty::serialize(e, "IntegerLiteral");
}
inline std::string serialize(const FloatLiteral &e) {
  return ty::serialize(e, "FloatLiteral");
}
inline std::string serialize(const BoolLiteral &e) {
  return ty::serialize(e, "BoolLiteral");
}
inline std::string serialize(const StringLiteral &e) {
  return ty::serialize(e, "StringLiteral");
}
inline std::string serialize(const Variable &e) {
  return R"({ "Variable": { "name": )" + ty::serialize(e.name) + " } }";
}
inline std::string serialize(const FunctionCall &e) {
  return R"({ "FunctionCall": { "function": )" + ty::serialize(e.function) +
         R"(, "argument": )" + ty::serialize(e.argument) + " } }";
}
inline std::string serialize(const TupleExpression &e) {
  return R"({ "TupleExpression": { "elements": )" + ty::serialize(e.elements) +
         " } }";
}
inline std::string serialize(const UnaryExpression &e) {
  return R"({ "UnaryExpression": { "op": )" + ty::serialize(e.op) +
         R"(, "expression": )" + ty::serialize(e.expression) + " } }";
}
inline std::string serialize(const BinaryExpression &e) {
  return R"({ "BinaryExpression": { "op": )" + ty::serialize(e.op) +
         R"(, "left": )" + ty::serialize(e.left) + R"(, "right": )" +
         ty::serialize(e.right) + " } }";
}

inline std::string serialize(const VariableDeclaration &e) {
  return R"({ "VariableDeclaration": { "name": )" + ty::serialize(e.name) +
         R"(, "body": )" + ty::serialize(e.body) + " } }";
}
inline std::string serialize(const Assignment &e) {
  return R"({ "Assignment": { "target": )" + ty::serialize(e.target) +
         R"(, "op": )" + ty::serialize(e.op) + R"(, "body": )" +
         ty::serialize(e.body) + " } }";
}
inline std::string serialize(const BranchStatement &e) {
  return R"({ "BranchStatement": { "condition": )" +
         ty::serialize(e.condition) + R"(, "then": )" +
         ty::serialize(e.thenBlock) + R"(, "else": )" +
         ty::serialize(e.elseBlock) + " } }";
}

template <class T> std::string serialize(Box<T> box) {
  return std::format(R"({{ "Box": {} }})", ty::serialize(box.get()));
}

template <class... T> inline std::string serialize(const Union<T...> &u) {
  return std::visit([](const auto &element) { return ty::serialize(element); },
                    u.variant);
}

template <class T> std::string serialize(T object) {
  return object.serialize();
}

template <class T>
inline std::string serialize_capture(std::string name, std::string typeId,
                                     T value) {
  return std::format(
      R"({{ "name": "{}", "type": {{ "TypeId": "{}" }}, "value": {} }})", name,
      typeId, ty::serialize(value));
}

inline std::string serialize_closure(std::string captures, std::string id) {
  return std::format(R"({{ "Closure": {{ "captures": [{}], "id": "{}" }} }})",
                     captures, id);
}

// check that all types have operator==
static_assert(std::equality_comparable<ty::TypeType>);
static_assert(std::equality_comparable<ty::IntType>);
static_assert(std::equality_comparable<ty::FloatType>);
static_assert(std::equality_comparable<ty::BoolType>);
static_assert(std::equality_comparable<ty::StringType>);
static_assert(std::equality_comparable<ty::TupleType>);
static_assert(std::equality_comparable<ty::ListType>);
static_assert(std::equality_comparable<ty::FnType>);
static_assert(std::equality_comparable<ty::StructType>);
static_assert(std::equality_comparable<ty::UnionType>);
static_assert(std::equality_comparable<ty::TupleType>);
static_assert(std::equality_comparable<ty::Type>);
static_assert(std::equality_comparable<Box<ty::Type>>);

// check that all expressions have operator==
static_assert(std::equality_comparable<Box<ty::Expression>>);
} // namespace ty

inline ty::Type Type = ty::TypeType{};
inline ty::Type Int = ty::IntType{};
inline ty::Type Float = ty::FloatType{};
inline ty::Type Bool = ty::BoolType{};
inline ty::Type String = ty::StringType{};

inline struct List_ {
  ty::Type operator()(ty::Type element) const {
    return box(ty::ListType{.element = element});
  }
  std::string serialize() const { return R"({ "Function": "List" })"; }
} List;

inline struct Fn_ {
  ty::Type operator()(ty::Type argument, ty::Type returnType) const {
    return box(ty::FnType{.argument = argument, .returnType = returnType});
  }
  std::string serialize() const { return R"({ "Function": "Fn" })"; }
} Fn;

inline struct toFloat_ {
  float operator()(int n) const { return n; }
  std::string serialize() const { return R"({ "Function": "toFloat" })"; }
} toFloat;

inline ty::Type Expression = box(ty::StructType{.name = "Expression"});

inline struct panic_ {
  [[noreturn]]
  ty::Union<> operator()(ty::String message) const {
    std::cerr << "panic: " << message << std::endl;
    exit(1);
  }
  std::string serialize() const { return R"({ "Function": "panic" })"; }
} panic;

inline struct stringLiteral_ {
  ty::Expression operator()(ty::String str) const {
    return ty::StringLiteral{.value = str};
  }
  std::string serialize() const { return R"({ "Function": "stringLiteral" })"; }
} stringLiteral;

inline struct variable_ {
  ty::Expression operator()(ty::String name) const {
    return ty::Variable{.name = name};
  }
  std::string serialize() const { return R"({ "Function": "variable" })"; }
} variable;

inline struct binaryExpression_ {
  ty::Expression operator()(ty::String op, ty::Expression left,
                            ty::Expression right) const {
    return box(ty::BinaryExpression{.op = op, .left = left, .right = right});
  }
  std::string serialize() const {
    return R"({ "Function": "binaryExpression" })";
  }
} binaryExpression;

inline struct functionCall_ {
  ty::Expression operator()(ty::Expression function,
                            ty::Expression argument) const {
    return box(ty::FunctionCall{.function = function, .argument = argument});
  }
  std::string serialize() const { return R"({ "Function": "functionCall" })"; }
} functionCall;

inline struct printlnString_ {
  std::tuple<> operator()(ty::String str) const {
    std::cout << str << std::endl;
    return std::make_tuple();
  }
  std::string serialize() const { return R"({ "Function": "printlnString" })"; }
} printlnString;

inline struct len_ {
  int operator()(ty::String s) const { return s.length(); }
  std::string serialize() const { return R"({ "Function": "len" })"; }
} len;

inline struct subString_ {
  ty::String operator()(ty::String s, int start, int end) const {
    if (start < 0) {
      panic("subString: start < 0");
    }
    if (end > s.length()) {
      panic("subString: end > len");
    }
    if (end < start) {
      panic("subString: end < start");
    }
    return s.subString(start, end);
  }
  std::string serialize() const { return R"({ "Function": "subString" })"; }
} subString;

// isVariant for subset
template <typename... T, typename... V>
  requires(sizeof...(T) != 1)
ty::Union<T...> isVariant_(ty::Union<V...> _union) {
  bool found = (std::holds_alternative<T>(_union.variant) || ...);
  return found;
}

// isVariant for element
template <typename T, typename... V> bool isVariant_(ty::Union<V...> _union) {
  return std::holds_alternative<T>(_union.variant);
}

// getVariant for subset
template <typename... T, typename... V>
  requires(sizeof...(T) != 1)
ty::Union<T...> getVariant_(ty::Union<V...> _union) {
  std::optional<ty::Union<T...>> result;
  (
      [&] {
        if (std::holds_alternative<T>(_union.variant)) {
          result = std::get<T>(_union.variant);
        }
      }(),
      ...);
  if (result.has_value()) {
    return result.value();
  }
  panic("getVariant: variant mismatch");
}

// getVariant for element
template <typename T, typename... V> T getVariant_(ty::Union<V...> _union) {
  if (!std::holds_alternative<T>(_union.variant)) {
    panic("getVariant: variant mismatch");
  }
  return std::get<T>(_union.variant);
}

inline struct Union_ {
  ty::Type operator()(ty::List<ty::Type> variants) const {
    if (variants.size() == 1) {
      return variants[0];
    }
    // Nested unions are flattened
    ty::List<ty::Type> flat_variants;
    for (const auto &variant : variants) {
      if (std::holds_alternative<ty::Box<ty::UnionType>>(variant.variant)) {
        for (auto nested_variant :
             std::get<ty::Box<ty::UnionType>>(variant.variant).get().variants) {
          flat_variants = flat_variants.append(nested_variant);
        }
      } else {
        flat_variants = flat_variants.append(variant);
      }
    }
    // Deduplicate types.
    // This is very stupid and inefficient, but it works.
    ty::List<ty::Type> unique_variants;
    {
      std::set<std::string> unique_strings;
      for (const auto &variant : flat_variants) {
        std::string serialized = ty::serialize(variant);
        if (!unique_strings.contains(serialized)) {
          unique_strings.insert(serialized);
          unique_variants = unique_variants.append(variant);
        }
      }
    }
    return box(ty::UnionType{.variants = unique_variants});
  }
  std::string serialize() const { return R"({ "Function": "Union" })"; }
} Union;
