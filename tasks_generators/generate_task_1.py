import re
from random import random, randint, choice
from string import ascii_lowercase

def choice_pop(lst):
    elem = choice(lst)
    lst.remove(elem)
    return elem

def chance(coef=0.5):
    return random() > coef

args = ''
letters = list(ascii_lowercase)
params_count = randint(2, 3)
params = []
for i in range(params_count):
	letter = choice_pop(letters)
	params.append(letter)
	args += f'\tARG({i+1}, int, {letter});\n'

comparators = ('>', '<', '>=', '<=', '==', '!=')
operators = ('+', '-', '*', '/')
pow_chance = 0.5
max_coef = 20

def get_comp_part(params):
    params = list(params)
    comparison = []
    param = choice_pop(params)
    comparison.append(param)
    comparison.append(choice(comparators))
    if params_count == 1 or chance():
        comparison.append('0')
    else:
        param = choice_pop(params)
        comparison.append(param)
        if chance():
            comparison.extend([
                '&&',
                param, 
                choice(comparators),
                '0'])
    return ' '.join(comparison)

comp = []
for _ in range(2):
    comp.append(get_comp_part(params))

def get_expr_part(params):
    params = list(params)
    expr = []
    if chance():
        expr.extend([str(randint(2, max_coef)), choice(operators)])
    param = choice_pop(params)
    expr.append(param)
    if chance():
        expr.extend(['*', param])
    if params_count > 1 and chance():
        expr.extend([choice(operators), choice_pop(params)])
    return ' '.join(expr)

expr = []
for _ in range(3):
    expr.append(get_expr_part(params))

reg1 = re.compile(r'(\w) \* \1')
reg2 = re.compile(r'(.) \&\& (.)')
tex_comp = []
for s in comp:
    s = reg1.sub(r'\1^2', s)
    s = reg2.sub(r'\1, \2', s)
    tex_comp.append(s)

tex_expr = []
for s in expr:
    s = reg1.sub(r'\1^2', s)
    s = reg2.sub(r'\1, \2', s)
    tex_expr.append(s)

expr = (comp[0], expr[0], comp[1], expr[1], expr[2])

latex_doc = 'Вычислить значение функции:\\\\\n\\begin{equation*}f(x,y) =\\begin{cases}' + tex_expr[0] + ', & \\textrm{если ' + tex_comp[0] +',}\\\\' + tex_expr[1] + ', & \\textrm{если ' + tex_comp[1] + ',}\\\\' + tex_expr[2]  + ' & \\textrm{в остальных случаях.}\\end{cases},\\end{equation*}\\\\\nПри вычислении значения необходимо использовать условную операцию сравнения ?:'
print(latex_doc)

template = '''
#include <stdio.h>
#include <test_utils.h>

int main(int argc, char* argv[])
{{
{}
        printf("%d", {} ? {} : ( {} ? ({}) : ({}) ) );
	return 0;
}}'''.format(args, *expr)
print(template)
