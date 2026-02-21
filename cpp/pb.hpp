#pragma once

// headers used by Yune programs
#include <tuple>      // std::tuple, std::apply
#include <string>     // std::string
#include <vector>     // std::vector
#include <iostream>   // std::cout

// headers for this file
#include <iomanip>
#include <sstream>
#include <memory>
#include <string>
#include <utility>
#include <variant>

template <class T> using Box = std::shared_ptr<T>;
template <class T> Box<T> box(T value) {
  return std::make_shared<T>(std::move(value));
}

namespace ty {
  template <class... T>
  struct Union {
    // Create from element
    template <class U>
    Union(U element) : variant(element) {}

    // Create from subset
    template <class... U>
    Union(Union<U...> subset)
    : variant(std::visit([](auto element) constexpr { return element; },
                         subset)) {}

    std::variant<T...> variant;
  };

template<class F, class Return, class... Args>
concept FunctionLike =
  requires(F f, Args... args) {
    { f(std::forward<Args>(args)...) } -> std::convertible_to<Return>;
    { f.serialize() } -> std::convertible_to<std::string>;
  };
  
  // A serializable function class similar to std::function
  template <class Return, class... Args> struct Function {
    struct Concept {
      virtual ~Concept() = default;
      virtual Return operator()(Args &&...args) const = 0;
      virtual std::string serialize() const = 0;
    };
    template <class F>
    struct Model final : Concept {
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
    : self(std::make_shared<Model<F>>(std::move(function))) {
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

  struct Span {
    Span(int line, int column) : line(line), column(column) {}
    int line;
    int column;
  };

  struct TypeType {};
  struct IntType {};
  struct FloatType {};
  struct BoolType {};
  struct StringType {};
  struct TupleType;
  struct ListType;
  struct FnType;
  struct StructType;

  using Type = Union<TypeType, IntType, FloatType, BoolType, StringType,
                     Box<TupleType>, Box<ListType>, Box<FnType>, Box<StructType>>;

  struct TupleType {
    std::vector<Type> elements;
  };
  struct ListType {
    Type element;
  };
  struct FnType {
    Type argument;
    Type returnType;
  };
  struct StructType {
    std::string name;
  };

  template <class T>
  struct Literal {
    T value;
  };
  using IntegerLiteral = Literal<int>;
  using FloatLiteral = Literal<float>;
  using BoolLiteral = Literal<bool>;
  using StringLiteral = Literal<std::string>;
  struct TupleExpression;
  struct FunctionCall;

  using Expression =
      Union<IntegerLiteral, FloatLiteral, BoolLiteral, StringLiteral, Box<TupleExpression>>;

  struct Variable {
    std::string name;
  };
  struct FunctionCall {
    Expression function;
    Expression argument;
  };
  struct TupleExpression {
    std::vector<Expression> elements;
  };
  struct UnaryExpression {
    std::string op;
    Expression expression;
  };
  struct BinaryExpression {
    std::string op;
    Expression left;
    Expression right;
  };

  struct VariableDeclaration;
  struct Assignment;
  struct BranchStatement;

  using Statement = Union<Box<VariableDeclaration>, Box<Assignment>, Box<BranchStatement>>;

  using Block = std::vector<Statement>;
  
  struct VariableDeclaration {
    std::string name;
    Block body;
  };
  struct Assignment {
    Variable target;
    std::string op;
    Block body;
  };
  struct BranchStatement {
    Expression condition;
    Block thenBlock;
    Block elseBlock;
  };
  
  // Escape string to JSON literal.
  inline std::string serialize(const std::string &s) {
    std::ostringstream oss;
    oss << '"';

    for (unsigned char c : s) {
      switch (c) {
      case '"':  oss << "\\\""; break;
      case '\\': oss << "\\\\"; break;
      case '\b': oss << "\\b";  break;
      case '\f': oss << "\\f";  break;
      case '\n': oss << "\\n";  break;
      case '\r': oss << "\\r";  break;
      case '\t': oss << "\\t";  break;
      default:
        if (c < 0x20) {
          oss << "\\u"
          << std::hex << std::setw(4) << std::setfill('0')
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

  inline std::string serialize(const TypeType &) { return R"({ "Type", {} })"; };
  inline std::string serialize(const IntType &) { return R"({ "IntType": {} })"; }
  inline std::string serialize(const FloatType &) {
    return R"({ "FloatType": {} })";
  }
  inline std::string serialize(const BoolType &) { return R"({ "BoolType": {} })"; }
  inline std::string serialize(const StringType &) {
    return R"({ "StringType": {} })";
  }
  std::string serialize(const TupleType &t);
  std::string serialize(const ListType &t);
  std::string serialize(const FnType &t);
  std::string serialize(const StructType &t);
  template <class T> std::string serialize(std::vector<T> elements);
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
  
  template <class T>
  std::string serialize(std::vector<T> elements) {
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
    std::apply([&](auto&&... elements) {
      (([&]() {
        oss << ty::serialize(elements);
        if (i + 1 < sizeof...(T)) {
          oss << ", ";
        }
        i++;
      }()), ...);
    }, tuple);
    oss << "] } }";
    return oss.str();
  }
  
  inline std::string serialize(const TupleType &t) {
    return R"({ "TupleType": { "elements": )" + ty::serialize(t.elements) + " } }";
  }
  inline std::string serialize(const ListType &t) {
    return R"({ "ListType": { "element": )" + ty::serialize(t.element) + " } }";
  }
  inline std::string serialize(const FnType &t) {
    return R"({ "FnType": { "argument": )" + ty::serialize(t.argument) + R"(, "return": )" + ty::serialize(t.returnType) + " } }";
  }
  inline std::string serialize(const StructType &t) {
    return R"({ "StructType": { "name": )" + ty::serialize(t.name) + " } }";
  }

  template<class T>
  inline std::string serialize(const Literal<T>& literal, std::string name) {
    return R"({ ")" + name + R"(": { "value": )" + ty::serialize(literal.value) +  " } }";
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
    return R"({ "FunctionType": { "function": )" + ty::serialize(e.function) + R"(, "argument": )" + ty::serialize(e.argument) + " } }";
  }
  inline std::string serialize(const TupleExpression &e) {
    return R"({ "TupleExpression": { "elements": )" + ty::serialize(e.elements) + " } }";
  }
  inline std::string serialize(const UnaryExpression &e) {
    return R"({ "UnaryExpression": { "op": )" + ty::serialize(e.op) + R"(, "expression": )" + ty::serialize(e.expression) + " } }";
  }
  inline std::string serialize(const BinaryExpression &e) {
    return R"({ "BinaryExpression": { "op": )" + ty::serialize(e.op) + R"(, "left": )" + ty::serialize(e.left) + R"(, "right": )" + ty::serialize(e.right) + " } }";
  }

  inline std::string serialize(const VariableDeclaration &e) {
    return R"({ "VariableDeclaration": { "name": )" + ty::serialize(e.name) + R"(, "body": )" + ty::serialize(e.body) + " } }";    
  }
  inline std::string serialize(const Assignment &e) {
    return R"({ "Assignment": { "target": )" + ty::serialize(e.target) + R"(, "op": )" + ty::serialize(e.op) + R"(, "body": )" + ty::serialize(e.body) + " } }";        
  }
  inline std::string serialize(const BranchStatement &e) {
    return R"({ "BranchStatement": { "condition": )" + ty::serialize(e.condition) + R"(, "then": )" + ty::serialize(e.thenBlock) + R"(, "else": )" + ty::serialize(e.elseBlock) + " } }";            
  }
  
  template<class T>
  std::string serialize(Box<T> box) {
    return ty::serialize(*box.get());
  }

  template <class... T> inline std::string serialize(const Union<T...> &u) {
    return std::visit([](const auto& element) { return ty::serialize(element); }, u.variant);
  }

  template <class T> std::string serialize(T object) {
    return object.serialize();
  }
} // namespace ty

