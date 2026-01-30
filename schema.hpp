#include <vector>
#include <string>


struct Value {};

struct Type : Value {};
struct IntType : Type {};
struct Float : Type {};
struct BoolType : Type {};
struct StringType : Type {};
struct TupleType : Type {
  std::vector<Type> elements;
};
struct ListType : Type {
  Type element;
};
struct FnType : Type {
  Type argument;
  Type _return;
};
struct StructType : Type {
  std::string name;
};

struct Expression : Value {};
struct String : Expression {
  std::string value;
};
