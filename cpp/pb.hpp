#pragma once

// headers also used by Yune programs
#include <algorithm>
#include <iostream> // std::cout
#include <string>   // std::string
#include <tuple>    // std::tuple, std::apply
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

template <class T> constexpr Box_t<T> box_f(T *value) { return Box_t(value); }

template <class T> Box_t<T> box_f(T &&value) {
  return std::make_shared<std::decay_t<T>>(std::forward<T>(value));
}

template <class... T> struct Union_t;

template <class> struct is_union_t : std::false_type {};
template <class... Ts> struct is_union_t<Union_t<Ts...>> : std::true_type {};

template <class T>
concept not_union_t = !is_union_t<std::remove_cvref_t<T>>::value;

template <class... T> struct Union_t {
  // Create from element directly
  template <class U>
    requires not_union_t<U> && (std::same_as<std::remove_cvref_t<U>, T> || ...)
  constexpr Union_t(U &&element) : variant(std::forward<U>(element)) {}

  // Create from element using an intermediate class
  // Required to create a Union[String] from const char*, for example
  template <class U>
    requires not_union_t<U> && (std::constructible_from<T, U> || ...) &&
             (!(std::same_as<std::remove_cvref_t<U>, T> || ...))
  constexpr Union_t(U &&element) : variant(std::forward<U>(element)) {}

  // Create from subset
  template <class... U>
  constexpr Union_t(const Union_t<U...> &subset)
      : variant(std::visit(
            [](auto &&element) -> std::variant<T...> {
              return std::variant<T...>(element);
            },
            subset.variant)) {}

  bool operator==(const Union_t<T...> &other) const = default;

  std::variant<T...> variant;
};

// Specialization of Union for zero elements.
// This is not constructable in Yune, but required for certain type signatures.
template <> struct Union_t<> {
  bool operator==(const Union_t<> &other) const = default;
};

template <class T> using List_t = std::vector<T>;

using String_t = std::string;

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
  bool operator==(const UnionType_t &other) const {
    if (variants.size() != other.variants.size()) {
      return false;
    }
    for (const auto &variant : variants) {
      if (std::find(other.variants.begin(), other.variants.end(), variant) ==
          other.variants.end()) {
        return false;
      }
    }
    return true;
  }
};

struct IntegerExpression_t {
  int location;
  int value;
};
struct FloatExpression_t {
  int location;
  float value;
};
struct BoolExpression_t {
  int location;
  bool value;
};
struct StringExpression_t {
  int location;
  String_t value;
};
struct VariableExpression_t {
  int location;
  String_t name;
};
struct FunctionCallExpression_t;
struct UnaryExpression_t;
struct BinaryExpression_t;
struct ListExpression_t;
struct TupleExpression_t;
struct ClosureExpression_t;
struct MacroExpression_t {
  int location;
  String_t macro;
  String_t text;
};
struct ValueExpression_t {
  int location;
  String_t json;
};

using Expression_t =
    Union_t<IntegerExpression_t, FloatExpression_t, BoolExpression_t,
            StringExpression_t, VariableExpression_t,
            Box_t<FunctionCallExpression_t>, Box_t<ListExpression_t>,
            Box_t<TupleExpression_t>, Box_t<ClosureExpression_t>,
            Box_t<MacroExpression_t>, Box_t<UnaryExpression_t>,
            Box_t<BinaryExpression_t>, ValueExpression_t>;

struct FunctionCallExpression_t {
  int location;
  Expression_t function;
  Expression_t argument;
};
struct UnaryExpression_t {
  int location;
  String_t op;
  Expression_t expression;
};
struct BinaryExpression_t {
  int location;
  String_t op;
  Expression_t left;
  Expression_t right;
};
struct ListExpression_t {
  int location;
  List_t<Expression_t> elements;
};
struct TupleExpression_t {
  int location;
  List_t<Expression_t> elements;
};

struct VariableDeclaration_t;
struct AssignStatement_t;
struct BranchStatement_t;
struct IsBranchStatement_t;
struct ExpressionStatement_t {
  Expression_t expression;
};

using Statement_t = Union_t<Box_t<VariableDeclaration_t>,
                            Box_t<AssignStatement_t>, Box_t<BranchStatement_t>,
                            Box_t<IsBranchStatement_t>, ExpressionStatement_t>;

using Block_t = List_t<Statement_t>;

struct VariableDeclaration_t {
  String_t name;
  Union_t<Expression_t, std::tuple<>> type;
  Block_t body;
};
struct AssignStatement_t {
  VariableExpression_t target;
  String_t op;
  Block_t body;
};
struct BranchStatement_t {
  Expression_t condition;
  Block_t thenBlock;
  Block_t elseBlock;
};
struct IsBranchStatement_t {
  Expression_t expression;
  String_t name;
  Expression_t type;
  Block_t thenBlock;
  Block_t elseBlock;
};

struct ClosureExpression_t {
  int location;
  List_t<std::tuple<String_t, Expression_t>> parameters;
  Expression_t returnType;
  Block_t body;
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
template <class T> std::string toJson_(List_t<T> list);
template <class... T> std::string toJson_(std::tuple<T...> tuple);
template <class... T> std::string toJson_(const Union_t<T...> &_union);
std::string toJson_(const IntegerExpression_t &e);
std::string toJson_(const FloatExpression_t &e);
std::string toJson_(const BoolExpression_t &e);
std::string toJson_(const StringExpression_t &e);
std::string toJson_(const VariableExpression_t &e);
std::string toJson_(const FunctionCallExpression_t &e);
std::string toJson_(const UnaryExpression_t &e);
std::string toJson_(const BinaryExpression_t &e);
std::string toJson_(const MacroExpression_t &e);
std::string toJson_(const ListExpression_t &e);
std::string toJson_(const TupleExpression_t &e);
std::string toJson_(const ClosureExpression_t &e);
inline std::string toJson_(const ValueExpression_t &e) {
  return std::format(
      R"({{ "ValueExpression": {{ "location": {}, "value": {} }} }})",
      toJson_(e.location), e.json);
}
std::string toJson_(const VariableDeclaration_t &e);
std::string toJson_(const AssignStatement_t &e);
std::string toJson_(const BranchStatement_t &e);
std::string toJson_(const ExpressionStatement_t &e);
// Fallback for classes that have a toJson_() method.
template <class T> std::string toJson_(T object);

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
  return array;
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
  return std::format(
      R"({{ "FnType": {{ "argument": {}, "returnType": {} }} }})",
      toJson_(t.argument), toJson_(t.returnType));
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

inline std::string toJson_(const IntegerExpression_t &e) {
  return std::format(
      R"({{ "IntegerExpression": {{ "location": {}, "value": {} }} }})",
      toJson_(e.location), toJson_(e.value));
}
inline std::string toJson_(const FloatExpression_t &e) {
  return std::format(
      R"({{ "FloatExpression": {{ "location": {}, "value": {} }} }})",
      toJson_(e.location), toJson_(e.value));
}
inline std::string toJson_(const BoolExpression_t &e) {
  return std::format(
      R"({{ "BoolExpression": {{ "location": {}, "value": {} }} }})",
      toJson_(e.location), toJson_(e.value));
}
inline std::string toJson_(const StringExpression_t &e) {
  return std::format(
      R"({{ "StringExpression": {{ "location": {}, "value": {} }} }})",
      toJson_(e.location), toJson_(e.value));
}
inline std::string toJson_(const VariableExpression_t &e) {
  return std::format(
      R"({{ "VariableExpression": {{ "location": {}, "name": {} }} }})",
      toJson_(e.location), toJson_(e.name));
}
inline std::string toJson_(const FunctionCallExpression_t &e) {
  return std::format(
      R"({{ "FunctionCallExpression": {{ "location": {}, "function": {}, "argument": {} }} }})",
      toJson_(e.location), toJson_(e.function), toJson_(e.argument));
}
inline std::string toJson_(const UnaryExpression_t &e) {
  return std::format(
      R"({{ "UnaryExpression": {{ "location": {}, "op": {}, "expression": {} }} }})",
      toJson_(e.location), toJson_(e.op), toJson_(e.expression));
}
inline std::string toJson_(const BinaryExpression_t &e) {
  return std::format(
      R"({{ "BinaryExpression": {{ "location": {}, "op": {}, "left": {}, "right": {} }} }})",
      toJson_(e.location), toJson_(e.op), toJson_(e.left), toJson_(e.right));
}
inline std::string toJson_(const ListExpression_t &e) {
  return std::format(
      R"({{ "ListExpression": {{ "location": {}, "elements": {} }} }})",
      toJson_(e.location), toJson_(e.elements));
}
inline std::string toJson_(const TupleExpression_t &e) {
  return std::format(
      R"({{ "TupleExpression": {{ "location": {}, "elements": {} }} }})",
      toJson_(e.location), toJson_(e.elements));
}
inline std::string toJson_(const ClosureExpression_t &e) {
  return std::format(
      R"({{ "ClosureExpression": {{ "location": {}, "parameters": {}, "returnType": {}, "body": {} }} }})",
      toJson_(e.location), toJson_(e.parameters), toJson_(e.returnType),
      toJson_(e.body));
}
inline std::string toJson_(const MacroExpression_t &e) {
  return std::format(
      R"({{ "MacroExpression": {{ "location": {}, "macro": {}, "text": {} }} }})",
      toJson_(e.location), toJson_(e.macro), toJson_(e.text));
}
inline std::string toJson_(const VariableDeclaration_t &e) {
  return std::format(
      R"({{ "VariableDeclaration": {{ "name": {}, "type": {}, "body": {} }} }})",
      toJson_(e.name), toJson_(e.type), toJson_(e.body));
}
inline std::string toJson_(const AssignStatement_t &e) {
  return R"({ "AssignStatement": { "target": )" + toJson_(e.target) +
         R"(, "op": )" + toJson_(e.op) + R"(, "body": )" + toJson_(e.body) +
         " } }";
}
inline std::string toJson_(const BranchStatement_t &e) {
  return R"({ "BranchStatement": { "condition": )" + toJson_(e.condition) +
         R"(, "then": )" + toJson_(e.thenBlock) + R"(, "else": )" +
         toJson_(e.elseBlock) + " } }";
}
inline std::string toJson_(const IsBranchStatement_t &e) {
  return std::format(
      R"({{ "IsBranchStatement": {{ "expression": {}, "name": {}, "type": {}, "then": {}, "else": {} }} }})",
      toJson_(e.expression), toJson_(e.name), toJson_(e.type),
      toJson_(e.thenBlock), toJson_(e.elseBlock));
}
inline std::string toJson_(const ExpressionStatement_t &e) {
  return std::format(R"({{ "ExpressionStatement": {{ "expression": {} }} }})",
                     toJson_(e.expression));
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
inline std::string capture_toJson_(std::string name, std::string typeId,
                                   T value) {
  return std::format(
      R"({{ "name": "{}", "type": {{ "TypeId": "{}" }}, "value": {} }})", name,
      typeId, toJson_(value));
}

inline std::string closure_toJson_(std::string captures, std::string id) {
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

inline Type_t Type = TypeType_t{};
inline Type_t Int = IntType_t{};
inline Type_t Float = FloatType_t{};
inline Type_t Bool = BoolType_t{};
inline Type_t String = StringType_t{};

inline Type_t Expression = box_f(StructType_t{.name = "Expression"});
inline Type_t Statement = box_f(StructType_t{.name = "Statement"});

inline struct List_f {
  Type_t operator()(Type_t element) const {
    return box_f(ListType_t{.element = element});
  }
  std::string toJson_() const { return R"({ "Function": "List" })"; }
} List;

inline struct Fn_f {
  Type_t operator()(Type_t argument, Type_t returnType) const {
    return box_f(FnType_t{.argument = argument, .returnType = returnType});
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

inline struct integerExpression_f {
  Expression_t operator()(int location, int value) const {
    return IntegerExpression_t{.location = location, .value = value};
  }
  std::string toJson_() const {
    return R"({ "Function": "integerExpression" })";
  }
} integerExpression;

inline struct floatExpression_f {
  Expression_t operator()(int location, float value) const {
    return FloatExpression_t{.location = location, .value = value};
  }
  std::string toJson_() const { return R"({ "Function": "floatExpression" })"; }
} floatExpression;

inline struct boolExpression_f {
  Expression_t operator()(int location, bool value) const {
    return BoolExpression_t{.location = location, .value = value};
  }
  std::string toJson_() const { return R"({ "Function": "boolExpression" })"; }
} boolExpression;

inline struct stringExpression_f {
  Expression_t operator()(int location, String_t value) const {
    return StringExpression_t{.location = location, .value = value};
  }
  std::string toJson_() const {
    return R"({ "Function": "stringExpression" })";
  }
} stringExpression;

inline struct variableExpression_f {
  Expression_t operator()(int location, String_t name) const {
    return VariableExpression_t{.location = location, .name = name};
  }
  std::string toJson_() const {
    return R"({ "Function": "variableExpression" })";
  }
} variableExpression;

inline struct unaryExpression_ {
  Expression_t operator()(int location, String_t op,
                          Expression_t expression) const {
    if (op != ";" && op != "-") {
      panic(std::format("Invalid unary operator: '{}'", op));
    }
    return box_f(UnaryExpression_t{
        .location = location, .op = op, .expression = expression});
  }
  std::string toJson_() const { return R"({ "Function": "unaryExpression" })"; }
} unaryExpression;

inline struct binaryExpression_ {
  Expression_t operator()(int location, String_t op, Expression_t left,
                          Expression_t right) const {
    if (op != "+" && op != "-" && op != "*" && op != "/" && op != "<" &&
        op != ">") {
      panic(std::format("Invalid binary operator: '{}'", op));
    }
    return box_f(BinaryExpression_t{
        .location = location, .op = op, .left = left, .right = right});
  }
  std::string toJson_() const {
    return R"({ "Function": "binaryExpression" })";
  }
} binaryExpression;

inline struct functionCallExpression_f {
  Expression_t operator()(int location, Expression_t function,
                          Expression_t argument) const {
    return box_f(FunctionCallExpression_t{
        .location = location, .function = function, .argument = argument});
  }
  std::string toJson_() const {
    return R"({ "Function": "functionCallExpression" })";
  }
} functionCallExpression;

inline struct expressionStatement_f {
  Statement_t operator()(Expression_t expression) const {
    return ExpressionStatement_t{.expression = expression};
  }
  std::string toJson_() const {
    return R"({ "Function": "expressionStatementExpression" })";
  }
} expressionStatement;

inline struct closureExpression_f {
  Expression_t operator()(int location,
                          List_t<std::tuple<String_t, Expression_t>> parameters,
                          Expression_t returnType, Block_t body) const {
    return box_f(ClosureExpression_t{.location = location,
                                     .parameters = parameters,
                                     .returnType = returnType,
                                     .body = body});
  }
  std::string toJson_() const {
    return R"({ "Function": "closureExpression" })";
  }
} closureExpression;

inline struct macroExpression_f {
  Expression_t operator()(int location, String_t macro, String_t text) const {
    return box_f(
        MacroExpression_t{.location = location, .macro = macro, .text = text});
  }
  std::string toJson_() const { return R"({ "Function": "macroExpression" })"; }
} macroExpression;

inline struct listExpression_f {
  Expression_t operator()(int location, List_t<Expression_t> elements) const {
    return box_f(ListExpression_t{.location = location, .elements = elements});
  }
  std::string toJson_() const { return R"({ "Function": "listExpression" })"; }
} listExpression;

inline struct tupleExpression_f {
  Expression_t operator()(int location, List_t<Expression_t> elements) const {
    return box_f(TupleExpression_t{.location = location, .elements = elements});
  }
  std::string toJson_() const { return R"({ "Function": "tupleExpression" })"; }
} tupleExpression;

inline struct variableDeclaration_f {
  Statement_t operator()(String_t name,
                         Union_t<Expression_t, std::tuple<>> type,
                         Block_t body) const {
    return box_f(
        VariableDeclaration_t{.name = name, .type = type, .body = body});
  }
  std::string toJson_() const {
    return R"({ "Function": "variableDeclaration" })";
  }
} variableDeclaration;

inline struct printlnString_f {
  std::tuple<> operator()(String_t str) const {
    std::cout << str << std::endl;
    return std::make_tuple();
  }
  std::string toJson_() const { return R"({ "Function": "printlnString" })"; }
} printlnString;

inline struct len_f {
  template <class T = int>
  int operator()(Union_t<String_t, List_t<T>> u) const {
    if (std::holds_alternative<String_t>(u)) {
      String_t s = std::get<String_t>(u);
      return s.length();
    } else {
      auto l = std::get<List_t<T>>(u);
      return l.length();
    }
  }
  template <class T = int> int operator()(List_t<T> l) const {
    return l.size();
  }
  int operator()(String_t s) const { return s.length(); }

  std::string toJson_() const { return R"({ "Function": "len" })"; }
} len;

inline struct get_f {
  template <class T = int> T operator()(List_t<T> list, int index) const {
    if (index < 0 || index >= list.size()) {
      panic("get: list index out of bounds");
    }
    return list[index];
  }

  std::string toJson_() const { return R"({ "Function": "get" })"; }
} get;

inline struct set_f {
  template <class T = int>
  std::tuple<> operator()(List_t<T> list, int index, T element) const {
    if (index < 0 || index >= list.size()) {
      panic("set: list index out of bounds");
    }
    list[index] = element;
  }

  std::string toJson_() const { return R"({ "Function": "set" })"; }
} set;

inline struct append_f {
  template <class T = int>
  List_t<T> operator()(List_t<T> list, T element) const {
    list.push_back(element);
    return list;
  }
  std::string toJson_() const { return R"({ "Function": "append" })"; }
} append;

inline struct subString_f {
  String_t operator()(String_t s, int start, int end) const {
    if (start < 0) {
      panic(std::format("subString: start ({}) < 0", start));
    }
    if (end > s.length()) {
      panic(std::format("subString: end ({}) > len ({})", end, s.length()));
    }
    if (end < start) {
      panic(std::format("subString: end ({}) < start ({})", end, start));
    }
    return s.substr(start, end - start);
  }
  std::string toJson_() const { return R"({ "Function": "subString" })"; }
} subString;

template <typename U, typename... T>
bool isSubset_(U _union) {
  bool found = (std::holds_alternative<T>(_union.variant) || ...);
  return found;
}

template <typename U, typename T> bool isVariant_(U _union) {
  return std::holds_alternative<T>(_union.variant);
}

template <typename U, typename... T>
Union_t<T...> getSubset_(U _union) {
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
  panic("getSubset: variant mismatch");
}

template <typename U, typename T> T getVariant_(U _union) {
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
          flat_variants.push_back(nested_variant);
        }
      } else {
        flat_variants.push_back(variant);
      }
    }
    // Deduplicate types.
    List_t<Type_t> unique_variants({});
  outer:
    for (const auto &variant : flat_variants) {
      for (const auto &unique_variant : unique_variants) {
        if (unique_variant == variant)
          goto outer;
      }
      unique_variants.push_back(variant);
    }
    return box_f(UnionType_t{.variants = unique_variants});
  }
  std::string toJson_() const { return R"({ "Function": "Union" })"; }
} Union;

template <class T> Expression_t inject(T value) {
  return ValueExpression_t{.json = toJson_(value)};
}
