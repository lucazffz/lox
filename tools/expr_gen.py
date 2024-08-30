# NOTE: script heavliy dependent on project structure

import subprocess
import argparse

parser = argparse.ArgumentParser(
    prog="expr_gen",
    description="Generates a file named expr.go to the"
    + " specified directory containing structs"
    + " used by the parser to produce an AST.",
)
parser.add_argument("output")

args = parser.parse_args()
out_dir = args.output


def define_type(file, class_name, fields=""):
    file.write("type " + class_name + " struct {\n")
    for field in fields.split(", "):
        if len(field) > 1:
            name = field.split(" ")[1]
            type = field.split(" ")[0]
            file.write(name + " " + type + ";")
            file.write("\n")
    file.write("}\n\n")


def define_ast(output_dir, base_name, *types):
    path = output_dir + "/" + base_name + ".go"
    with open(path, "w+") as file:
        file.write("// Code generated by tools/expr_gen.py. DO NOT EDIT.\n\n")
        file.write("package ast\n\n")
        file.write('import "github.com/LucazFFz/lox/internal/token"\n\n')
        file.write("type Expr interface {\nPrettyPrint\n}\n\n")
        for type in types:
            if type.find(":") != -1:
                class_name = type.split(":")[0].strip()
                fields = type.split(":")[1].strip()
                define_type(file, class_name, fields)
            else:
                define_type(file, type)


define_ast(
    out_dir,
    "expr",
    "Binary : Expr Left, token.Token Op, Expr Right",
    "Grouping : Expr Expr",
    "Literal : string Value",
    "Unary : token.Token Op, Expr Right",
    "Ternary : Expr Condition, Expr Left, Expr Right",
    "Nothing",
)

subprocess.run(["go", "fmt", out_dir + "/expr.go"])
