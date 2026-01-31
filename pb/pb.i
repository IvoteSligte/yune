%module pb
%{
#include "../cpp/pb.hpp"
%}
%include <std_vector.i>
%include <std_string.i>

%template(TypeVector) std::vector<pb::Type>;
%template(ValueVector) std::vector<pb::Value>;

%rename(eq) operator==;

%include "../cpp/pb.hpp"
