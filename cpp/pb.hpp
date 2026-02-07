
#include <iomanip>
#include <sstream>
#include <memory>
#include <string>
#include <utility>
#include <variant>
#include <vector>

template <class T> using Box = std::shared_ptr<T>;
template <class T> Box<T> box(T value) {
  return std::make_shared<T>(std::move(value));
}

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
    Union(U element) : variant(element) {}

    // Create from subset
    template <class... U>
    Union(Union<U...> subset)
    : variant(std::visit([](auto element) constexpr { return element; },
                         subset)) {}

    std::variant<T...> variant;
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

  using Expression =
      Union<IntegerLiteral, FloatLiteral, BoolLiteral, StringLiteral, Box<TupleExpression>>;

  struct TupleExpression {
    std::vector<Expression> elements;
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
  template<class T> std::string serialize(const Literal<T>& literal, std::string name);
  
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
    oss << '[';
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
    oss << ']';
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
  
  template <class... T>
  inline std::string serialize(const TupleExpression &e) {
    return R"({ "TupleExpression": { "elements": )" + ty::serialize(e) + " } }";
  }

  // TODO: other expression kinds

  template<class T>
  std::string serialize(Box<T> box) {
    return ty::serialize(*box.get());
  }

  template <class... T> inline std::string serialize(const Union<T...> &u) {
    return std::visit([](const auto& element) { return ty::serialize(element); }, u.variant);
  }
} // namespace ty

