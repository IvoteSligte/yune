%module pb
%{
#include "pb.h"
%}
%include <std_vector.i>
%include <std_string.i>

%template(TypeVector) std::vector<Type>;
 
%include "pb.h"
