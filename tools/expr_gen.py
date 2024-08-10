import subprocess
import argparse

parser = argparse.ArgumentParser(prog = "expr_gen", description= "Generates a file named expr.go to the specified directory containing ast structs.")
parser.add_argument("output")

args = parser.parse_args()
out_dir = args.output

def define_type(file, class_name, fields):
    file.write("type " + class_name + " struct {\n")
    for field in fields.split(", "):
        name = field.split(" ")[1]
        type = field.split(" ")[0]
        
        file.write(name + " " + type + ";")
        file.write("\n")
    file.write("}\n\n")

def define_ast(output_dir, base_name, *types):
    path = output_dir + "/" + base_name + ".go"
    with open(path, "w+") as file:
        file.write("package ast\n\n")
        file.write("import \"github.com/LucazFFz/lox/internal/token\"\n\n") 
        file.write("type Expr interface {\nPrettyPrint\n}\n\n")
        for type in types:
            class_name = type.split(":")[0].strip()
            fields = type.split(":")[1].strip()
            define_type(file, class_name, fields)
            

define_ast(out_dir, "expr", 
           "Binary : Expr left, token.Token op, Expr right", 
           "Grouping : Expr expr", 
           "Literal : byte value", 
           "Unary : token.Token op, Expr right")

subprocess.run(["go", "fmt", out_dir + "/expr.go"])


