only execute globals initialized with pure functions at compile-time

clang++ -shared -fPIC mylib.cpp -o libmylib.so
