package ast

type Value interface {
    GetType() InferredType
}
