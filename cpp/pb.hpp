#pragma once

// headers also used by Yune programs
#include <algorithm>
#include <iostream> // std::cout
#include <set>
#include <string> // std::string
#include <tuple>  // std::tuple, std::apply
#include <type_traits>
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

template <class T> struct Box_t {
  constexpr Box_t(T *ptr) : ptr(ptr) {}

  Box_t(T &&value)
      : ptr(std::make_shared<std::decay_t<T>>(std::forward<T>(value))) {}

  Box_t(std::shared_ptr<T> &&ptr) : ptr(ptr) {}

  bool operator==(const Box_t<T> &other) const {
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

template <class T> constexpr Box_t<T> box(T *value) { return Box_t(value); }

template <class T> Box_t<T> box(T &&value) {
  return std::make_shared<std::decay_t<T>>(std::forward<T>(value));
}

template <class... T> struct Union_t {
  // Create from element directly
  template <class U>
    requires(std::is_same_v<std::decay_t<U>, T> || ...) // U is one of T
  constexpr Union_t(U &&element) : variant(std::forward<U>(element)) {}

  // Create from element using an intermediate class
  // Required to create a Union[String] from a const char*, for example
  template <class U>
    requires(std::is_constructible_v<U, T> ||
             ...) && // any T can be constructed from U
            (!(std::is_same_v<std::decay_t<U>, T> || ...)) // U is not one of T
  constexpr Union_t(U &&element) : variant(std::forward<U>(element)) {}

  // Create from subset
  template <class... U>
  constexpr Union_t(const Union_t<U...> &subset)
      : variant(std::visit([](auto &&element) constexpr
                               -> std::variant<T...> { return element; },
                           subset.variant)) {}

  bool operator==(const Union_t<T...> &other) const = default;

  std::variant<T...> variant;
};

// Specialization of Union for zero elements.
// This is not constructable in Yune, but required for certain type signatures.
template <> struct Union_t<> {
  bool operator==(const Union_t<> &other) const = default;
};

template <class T> struct List_t {
  struct ArrayRef {
    size_t size;
    const T *ptr;
  };

  // Prevent valueless std::variant
  constexpr List_t() : value(ArrayRef{.size = 0, .ptr = nullptr}) {}

  constexpr List_t(const T *array, size_t size)
      : value(ArrayRef{.size = size, .ptr = array}) {}

  List_t(std::initializer_list<T> value) : value(std::vector(value)) {}
  List_t(std::vector<T> value) : value(value) {}

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
  List_t<T> append(T element) const {
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

  const T *begin() const {
    if (std::holds_alternative<std::vector<T>>(value)) {
      return std::get<std::vector<T>>(value).data();
    } else {
      return std::get<ArrayRef>(value).ptr;
    }
  }

  const T *end() const {
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

  bool operator==(const List_t<T> &other) const {
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
struct String_t {
  constexpr String_t(const char *string) : value(string) {}
  constexpr String_t(std::string string) : value(string) {}

  std::string to_owned() const { return std::string(begin(), end()); }

  size_t length() const {
    if (std::holds_alternative<std::string>(value)) {
      return std::get<std::string>(value).length();
    } else {
      return std::string(std::get<const char *>(value)).length();
    }
  }

  String_t subString(int start, int end) const {
    if (std::holds_alternative<std::string>(value)) {
      return std::get<std::string>(value).substr(start, end - start);
    } else {
      return std::string(std::get<const char *>(value), start, end - start);
    }
  }

  String_t operator+(const String_t &other) const {
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

  bool operator==(const String_t &other) const {
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

inline std::ostream &operator<<(std::ostream &os, const String_t &s) {
  for (const char c : s) {
    os << c;
  }
  return os;
}

template <class F, class Return, class... Args>
concept FunctionLike_ = requires(F f, Args... args) {
  { f(std::forward<Args>(args)...) } -> std::convertible_to<Return>;
  { f.toJson_() } -> std::convertible_to<std::string>;
};

// FIXME: this is probably not runtime-free
// A serializable function class similar to std::function
template <class Return, class... Args> struct Fn_t {
  struct Concept {
    virtual ~Concept() = default;
    virtual Return operator()(Args &&...args) const = 0;
    virtual std::string toJson_() const = 0;
  };
  template <class F> struct Model final : Concept {
    explicit Model(F f) : function(std::move(f)) {}

    Return operator()(Args &&...args) const override {
      return function(std::forward<Args>(args)...);
    }
    std::string toJson_() const override { return function.toJson_(); }

    F function;
  };
  template <class F>
    requires FunctionLike_<F, Return, Args...>
  Fn_t(F function)
      : self(std::make_shared<std::decay_t<Model<F>>>(std::move(function))) {
    static_assert(std::is_class_v<F>,
                  "Function requires callable object, not function pointer");
  }

  Return operator()(Args... args) const {
    return (*self)(std::forward<Args>(args)...);
  }
  std::string toJson_() const { return self->toJson_(); }
  // TODO: copy and move operators

  std::shared_ptr<Concept> self;
};

// extends std::apply to work for a zero-sized tuple
template <class F, class Tuple> decltype(auto) apply_(F &&f, Tuple &&tuple) {
  if constexpr (std::tuple_size_v<std::remove_reference_t<Tuple>> == 0) {
    return std::forward<F>(f)();
  } else {
    return std::apply(std::forward<F>(f), std::forward<Tuple>(tuple));
  }
}

struct Span_t {
  Span_t(int line, int column) : line(line), column(column) {}
  int line;
  int column;
};

struct TypeType_t {
  bool operator==(const TypeType_t &other) const { return true; }
};
struct IntType_t {
  bool operator==(const IntType_t &other) const { return true; }
};
struct FloatType_t {
  bool operator==(const FloatType_t &other) const { return true; }
};
struct BoolType_t {
  bool operator==(const BoolType_t &other) const { return true; }
};
struct StringType_t {
  bool operator==(const StringType_t &other) const { return true; }
};
struct TupleType_t;
struct ListType_t;
struct FnType_t;
struct StructType_t;
struct UnionType_t;

using Type_t =
    Union_t<TypeType_t, IntType_t, FloatType_t, BoolType_t, StringType_t,
            Box_t<TupleType_t>, Box_t<ListType_t>, Box_t<FnType_t>,
            Box_t<StructType_t>, Box_t<UnionType_t>>;

struct TupleType_t {
  List_t<Type_t> elements;
  bool operator==(const TupleType_t &other) const = default;
};
struct ListType_t {
  Type_t element;
  bool operator==(const ListType_t &other) const = default;
};
struct FnType_t {
  Type_t argument;
  Type_t returnType;
  bool operator==(const FnType_t &other) const = default;
};
struct StructType_t {
  struct Field {
    String_t name;
    Type_t type;
    bool operator==(const Field &other) const = default;
  };

  String_t name;
  List_t<Field> fields;
  bool operator==(const StructType_t &other) const = default;
};
struct UnionType_t {
  List_t<Type_t> variants;
  bool operator==(const UnionType_t &other) const = default;
};

namespace { // hidden
template <class T> struct Literal {
  T value;
};
} // namespace

using IntegerLiteral_t = Literal<int>;
using FloatLiteral_t = Literal<float>;
using BoolLiteral_t = Literal<bool>;
using StringLiteral_t = Literal<String_t>;
struct Variable_t;
struct FunctionCall_t;
struct TupleExpression_t;
struct UnaryExpression_t;
struct BinaryExpression_t;

using Expression_t =
    Union_t<IntegerLiteral_t, FloatLiteral_t, BoolLiteral_t, StringLiteral_t,
            Variable_t, Box_t<FunctionCall_t>, Box_t<TupleExpression_t>,
            Box_t<UnaryExpression_t>, Box_t<BinaryExpression_t>>;

struct Variable_t {
  String_t name;
};
struct FunctionCall_t {
  Expression_t function;
  Expression_t argument;
};
struct TupleExpression_t {
  List_t<Expression_t> elements;
};
struct UnaryExpression_t {
  String_t op;
  Expression_t expression;
};
struct BinaryExpression_t {
  String_t op;
  Expression_t left;
  Expression_t right;
};

struct VariableDeclaration_t;
struct Assignment_t;
struct BranchStatement_t;

using Statement_t = Union_t<Box_t<VariableDeclaration_t>, Box_t<Assignment_t>,
                            Box_t<BranchStatement_t>>;

using Block_t = List_t<Statement_t>;

struct VariableDeclaration_t {
  String_t name;
  Block_t body;
};
struct Assignment_t {
  Variable_t target;
  String_t op;
  Block_t body;
};
struct BranchStatement_t {
  Expression_t condition;
  Block_t thenBlock;
  Block_t elseBlock;
};

// Escape string to JSON literal.
inline std::string toJson_(const String_t &s) {
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
inline std::string toJson_(const int &i) { return std::to_string(i); }
inline std::string toJson_(const bool &b) { return std::to_string(b); }
inline std::string toJson_(const float &f) { return std::to_string(f); }

inline std::string toJson_(const TypeType_t &) {
  return R"({ "TypeType": {} })";
};
inline std::string toJson_(const IntType_t &) { return R"({ "IntType": {} })"; }
inline std::string toJson_(const FloatType_t &) {
  return R"({ "FloatType": {} })";
}
inline std::string toJson_(const BoolType_t &) {
  return R"({ "BoolType": {} })";
}
inline std::string toJson_(const StringType_t &) {
  return R"({ "StringType": {} })";
}
std::string toJson_(const TupleType_t &t);
std::string toJson_(const ListType_t &t);
std::string toJson_(const FnType_t &t);
std::string toJson_(const StructType_t &t);
std::string toJson_(const UnionType_t &t);
template <class T> std::string toJson_(List_t<T> elements);
template <class... T> std::string toJson_(std::tuple<T...> elements);
template <class... T> std::string toJson_(const Union_t<T...> &u);
std::string toJson_(const IntegerLiteral_t &e);
std::string toJson_(const FloatLiteral_t &e);
std::string toJson_(const BoolLiteral_t &e);
std::string toJson_(const StringLiteral_t &e);
std::string toJson_(const Variable_t &e);
std::string toJson_(const FunctionCall_t &e);
std::string toJson_(const TupleExpression_t &e);
std::string toJson_(const UnaryExpression_t &e);
std::string toJson_(const BinaryExpression_t &e);
std::string toJson_(const VariableDeclaration_t &e);
std::string toJson_(const Assignment_t &e);
std::string toJson_(const BranchStatement_t &e);
// Fallback for classes that have a toJson_() method.
template <class T> std::string toJson_(T object);

// Helper for converting a C++ type to JSON.
// Cannot simply be a function because partial specialization is required for
// TypeToJson_<List_t<T>>.
template <class T> struct TypeValueOf_ {
  static Type_t value();
};

template <> struct TypeValueOf_<int> {
  static Type_t value() { return IntType_t(); }
};

template <> struct TypeValueOf_<float> {
  static Type_t value() { return FloatType_t(); }
};

template <> struct TypeValueOf_<bool> {
  static Type_t value() { return BoolType_t(); }
};

template <> struct TypeValueOf_<String_t> {
  static Type_t value() { return StringType_t(); }
};

template <> struct TypeValueOf_<Type_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<TypeType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<IntType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<FloatType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<BoolType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<StringType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<TupleType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<ListType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<FnType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<StructType_t> {
  static Type_t value() { return TypeType_t(); }
};
template <> struct TypeValueOf_<StructType_t::Field> {
  static Type_t value() {
    std::cerr << "Tried to call TypeValueOf_<StructType_t::Field>::value()"
              << std::endl;
    abort();
  } // not sure why the linker wants this
};
template <> struct TypeValueOf_<UnionType_t> {
  static Type_t value() { return TypeType_t(); }
};

inline Type_t Expression = box(StructType_t{.name = "Expression"});
inline Type_t Statement = box(StructType_t{.name = "Statement"});

template <> struct TypeValueOf_<IntegerLiteral_t> {
  static Type_t value() { return Expression; }
};

template <class T> struct TypeValueOf_<Literal<T>> {
  static Type_t value() { return Expression; }
};

template <> struct TypeValueOf_<Variable_t> {
  static Type_t value() { return Expression; }
};
template <> struct TypeValueOf_<FunctionCall_t> {
  static Type_t value() { return Expression; }
};
template <> struct TypeValueOf_<TupleExpression_t> {
  static Type_t value() { return Expression; }
};
template <> struct TypeValueOf_<UnaryExpression_t> {
  static Type_t value() { return Expression; }
};
template <> struct TypeValueOf_<BinaryExpression_t> {
  static Type_t value() { return Expression; }
};
template <> struct TypeValueOf_<VariableDeclaration_t> {
  static Type_t value() { return Statement; }
};
template <> struct TypeValueOf_<Assignment_t> {
  static Type_t value() { return Statement; }
};
template <> struct TypeValueOf_<BranchStatement_t> {
  static Type_t value() { return Statement; }
};

template <class... T> struct TypeValueOf_<std::tuple<T...>> {
  static Type_t value() {
    return box(TupleType_t{
        .elements = {typeValueOf_<T>()...},
    });
  }
};

template <class T> struct TypeValueOf_<Box_t<T>> {
  static Type_t value() { return TypeValueOf_<T>::value(); }
};

template <class T> struct TypeValueOf_<List_t<T>> {
  static Type_t value() {
    return ListType_t{
        .element = typeValueOf_<T>(),
    };
  }
};

template <class Return, class Arg> struct TypeValueOf_<Fn_t<Return, Arg>> {
  static Type_t value() {
    return box(FnType_t{
        .argument = typeValueOf_<Arg>(),
        .returnType = typeValueOf_<Return>(),
    });
  }
};

template <class Return, class... Args>
  requires(sizeof...(Args) != 1)
struct TypeValueOf_<Fn_t<Return, Args...>> {
  static Type_t value() {
    return box(FnType_t{
        .argument = typeValueOf_<std::tuple<Args...>>(),
        .returnType = typeValueOf_<Return>(),
    });
  }
};

// NOTE: StructType_t not implemented because there is no way to match all
// struct types (I think)

template <class... Variants> struct TypeValueOf_<Union_t<Variants...>> {
  static Type_t value() {
    return box(UnionType_t{
        .variants = {TypeValueOf_<Variants>::value()...},
    });
  }
};

template <class T> std::string toJson_(List_t<T> list) {
  std::ostringstream oss;
  oss << '[';
  for (int i = 0; i < list.size(); i++) {
    oss << toJson_(list[i]);
    if (i + 1 < list.size()) {
      oss << ", ";
    }
  }
  oss << ']';
  std::string array = oss.str();

  return std::format(R"({{ "List": {}, "generic_": {} }})", array,
                     toJson_(TypeValueOf_<T>::value()));
}

template <class... T> std::string toJson_(std::tuple<T...> tuple) {
  std::ostringstream oss;
  oss << R"({ "Tuple": { "elements": [)";
  int i = 0;
  std::apply(
      [&](auto &&...elements) {
        (([&]() {
           oss << toJson_(elements);
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

inline std::string toJson_(const TupleType_t &t) {
  return R"({ "TupleType": { "elements": )" + toJson_(t.elements) + " } }";
}
inline std::string toJson_(const ListType_t &t) {
  return R"({ "ListType": { "element": )" + toJson_(t.element) + " } }";
}
inline std::string toJson_(const FnType_t &t) {
  return R"({ "FnType": { "argument": )" + toJson_(t.argument) +
         R"(, "return": )" + toJson_(t.returnType) + " } }";
}
inline std::string toJson_(const StructType_t::Field &t) {
  return std::format(R"({{ "name": {}, "type": {} }})", toJson_(t.name),
                     toJson_(t.type));
}
inline std::string toJson_(const StructType_t &t) {
  return std::format(R"({{ "StructType": {{ "name": {}, "fields": {} }} }})",
                     toJson_(t.name), toJson_(t.fields));
}
inline std::string toJson_(const UnionType_t &t) {
  return R"({ "UnionType": { "variants": )" + toJson_(t.variants) + " } }";
}

// -- expressions --

template <class T>
inline std::string toJson_(const Literal<T> &literal, std::string name) {
  return R"({ ")" + name + R"(": { "value": )" + toJson_(literal.value) +
         " } }";
}
inline std::string toJson_(const IntegerLiteral_t &e) {
  return toJson_(e, "IntegerLiteral");
}
inline std::string toJson_(const FloatLiteral_t &e) {
  return toJson_(e, "FloatLiteral");
}
inline std::string toJson_(const BoolLiteral_t &e) {
  return toJson_(e, "BoolLiteral");
}
inline std::string toJson_(const StringLiteral_t &e) {
  return toJson_(e, "StringLiteral");
}
inline std::string toJson_(const Variable_t &e) {
  return R"({ "Variable": { "name": )" + toJson_(e.name) + " } }";
}
inline std::string toJson_(const FunctionCall_t &e) {
  return R"({ "FunctionCall": { "function": )" + toJson_(e.function) +
         R"(, "argument": )" + toJson_(e.argument) + " } }";
}
inline std::string toJson_(const TupleExpression_t &e) {
  return R"({ "TupleExpression": { "elements": )" + toJson_(e.elements) +
         " } }";
}
inline std::string toJson_(const UnaryExpression_t &e) {
  return R"({ "UnaryExpression": { "op": )" + toJson_(e.op) +
         R"(, "expression": )" + toJson_(e.expression) + " } }";
}
inline std::string toJson_(const BinaryExpression_t &e) {
  return R"({ "BinaryExpression": { "op": )" + toJson_(e.op) + R"(, "left": )" +
         toJson_(e.left) + R"(, "right": )" + toJson_(e.right) + " } }";
}

inline std::string toJson_(const VariableDeclaration_t &e) {
  return R"({ "VariableDeclaration": { "name": )" + toJson_(e.name) +
         R"(, "body": )" + toJson_(e.body) + " } }";
}
inline std::string toJson_(const Assignment_t &e) {
  return R"({ "Assignment": { "target": )" + toJson_(e.target) + R"(, "op": )" +
         toJson_(e.op) + R"(, "body": )" + toJson_(e.body) + " } }";
}
inline std::string toJson_(const BranchStatement_t &e) {
  return R"({ "BranchStatement": { "condition": )" + toJson_(e.condition) +
         R"(, "then": )" + toJson_(e.thenBlock) + R"(, "else": )" +
         toJson_(e.elseBlock) + " } }";
}

template <class T> std::string toJson_(Box_t<T> box) {
  return std::format(R"({{ "Box": {} }})", toJson_(box.get()));
}

template <class... T> inline std::string toJson_(const Union_t<T...> &u) {
  return std::visit([](const auto &element) { return toJson_(element); },
                    u.variant);
}

template <class T> std::string toJson_(T object) { return object.toJson_(); }

template <class T>
inline std::string serialize_capture_(std::string name, std::string typeId,
                                      T value) {
  return std::format(
      R"({{ "name": "{}", "type": {{ "TypeId": "{}" }}, "value": {} }})", name,
      typeId, toJson_(value));
}

inline std::string serialize_closure_(std::string captures, std::string id) {
  return std::format(R"({{ "Closure": {{ "captures": [{}], "id": "{}" }} }})",
                     captures, id);
}

// check that all types have operator==
static_assert(std::equality_comparable<TypeType_t>);
static_assert(std::equality_comparable<IntType_t>);
static_assert(std::equality_comparable<FloatType_t>);
static_assert(std::equality_comparable<BoolType_t>);
static_assert(std::equality_comparable<StringType_t>);
static_assert(std::equality_comparable<TupleType_t>);
static_assert(std::equality_comparable<ListType_t>);
static_assert(std::equality_comparable<FnType_t>);
static_assert(std::equality_comparable<StructType_t>);
static_assert(std::equality_comparable<UnionType_t>);
static_assert(std::equality_comparable<TupleType_t>);
static_assert(std::equality_comparable<Type_t>);
static_assert(std::equality_comparable<Box_t<Type_t>>);

// check that all expressions have operator==
static_assert(std::equality_comparable<Box_t<Expression_t>>);

inline Type_t Type = TypeType_t{};
inline Type_t Int = IntType_t{};
inline Type_t Float = FloatType_t{};
inline Type_t Bool = BoolType_t{};
inline Type_t String = StringType_t{};

inline struct List_f {
  Type_t operator()(Type_t element) const {
    return box(ListType_t{.element = element});
  }
  std::string toJson_() const { return R"({ "Function": "List" })"; }
} List;

inline struct Fn_f {
  Type_t operator()(Type_t argument, Type_t returnType) const {
    return box(FnType_t{.argument = argument, .returnType = returnType});
  }
  std::string toJson_() const { return R"({ "Function": "Fn" })"; }
} Fn;

inline struct toFloat_f {
  float operator()(int n) const { return n; }
  std::string toJson_() const { return R"({ "Function": "toFloat" })"; }
} toFloat;

inline struct panic_f {
  [[noreturn]]
  Union_t<> operator()(String_t message) const {
    std::cerr << "panic: " << message << std::endl;
    exit(1);
  }
  std::string toJson_() const { return R"({ "Function": "panic" })"; }
} panic;

inline struct stringLiteral_f {
  Expression_t operator()(String_t str) const {
    return StringLiteral_t{.value = str};
  }
  std::string toJson_() const { return R"({ "Function": "stringLiteral" })"; }
} stringLiteral;

inline struct variable_f {
  Expression_t operator()(String_t name) const {
    return Variable_t{.name = name};
  }
  std::string toJson_() const { return R"({ "Function": "variable" })"; }
} variable;

inline struct binaryExpression_ {
  Expression_t operator()(String_t op, Expression_t left,
                          Expression_t right) const {
    return box(BinaryExpression_t{.op = op, .left = left, .right = right});
  }
  std::string toJson_() const {
    return R"({ "Function": "binaryExpression" })";
  }
} binaryExpression;

inline struct functionCall_f {
  Expression_t operator()(Expression_t function, Expression_t argument) const {
    return box(FunctionCall_t{.function = function, .argument = argument});
  }
  std::string toJson_() const { return R"({ "Function": "functionCall" })"; }
} functionCall;

inline struct printlnString_f {
  std::tuple<> operator()(String_t str) const {
    std::cout << str << std::endl;
    return std::make_tuple();
  }
  std::string toJson_() const { return R"({ "Function": "printlnString" })"; }
} printlnString;

inline struct len_f {
  int operator()(String_t s) const { return s.length(); }
  std::string toJson_() const { return R"({ "Function": "len" })"; }
} len;

inline struct subString_f {
  String_t operator()(String_t s, int start, int end) const {
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
  std::string toJson_() const { return R"({ "Function": "subString" })"; }
} subString;

// isVariant for subset
template <typename... T, typename... V>
  requires(sizeof...(T) != 1)
Union_t<T...> isVariant_(Union_t<V...> _union) {
  bool found = (std::holds_alternative<T>(_union.variant) || ...);
  return found;
}

// isVariant for element
template <typename T, typename... V> bool isVariant_(Union_t<V...> _union) {
  return std::holds_alternative<T>(_union.variant);
}

// getVariant for subset
template <typename... T, typename... V>
  requires(sizeof...(T) != 1)
Union_t<T...> getVariant_(Union_t<V...> _union) {
  std::optional<Union_t<T...>> result;
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
template <typename T, typename... V> T getVariant_(Union_t<V...> _union) {
  if (!std::holds_alternative<T>(_union.variant)) {
    panic("getVariant: variant mismatch");
  }
  return std::get<T>(_union.variant);
}

inline struct Union_f {
  Type_t operator()(List_t<Type_t> variants) const {
    if (variants.size() == 1) {
      return variants[0];
    }
    // Nested unions are flattened
    List_t<Type_t> flat_variants({});
    for (const auto &variant : variants) {
      if (std::holds_alternative<Box_t<UnionType_t>>(variant.variant)) {
        for (auto nested_variant :
             std::get<Box_t<UnionType_t>>(variant.variant).get().variants) {
          flat_variants = flat_variants.append(nested_variant);
        }
      } else {
        flat_variants = flat_variants.append(variant);
      }
    }
    // Deduplicate types.
    // This is very stupid and inefficient, but it works.
    List_t<Type_t> unique_variants({});
    {
      std::set<std::string> unique_strings;
      for (const auto &variant : flat_variants) {
        std::string serialized = ::toJson_(variant);
        if (!unique_strings.contains(serialized)) {
          unique_strings.insert(serialized);
          unique_variants = unique_variants.append(variant);
        }
      }
    }
    return box(UnionType_t{.variants = unique_variants});
  }
  std::string toJson_() const { return R"({ "Function": "Union" })"; }
} Union;
