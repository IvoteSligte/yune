%module pb
%{
#include "pb.hpp"
%}
%include <std_vector.i>
%include <std_string.i>

%template(TypeVector) std::vector<Type>;
%template(ValueVector) std::vector<Value>;

%rename(eq) operator==;

%include "pb.hpp"
