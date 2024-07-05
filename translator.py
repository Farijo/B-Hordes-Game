from string import Template
from os.path import basename
import argparse
import yaml

parser = argparse.ArgumentParser(
                    prog=basename(__file__),
                    description='Generate a translated file based on template template_file and translation file translation_file')
parser.add_argument('template_file')
parser.add_argument('translation_file')
parser.add_argument('-o', '--output', help='Write to OUTPUT instead of stdout')
args = parser.parse_args()

class GoTemplate(Template):
    delimiter = '@@'
    idpattern = '(?a:[\-_a-z][\-_a-z0-9]*)'

with open(args.template_file) as f:
    t = GoTemplate(f.read())

with open(args.translation_file) as f:
    y = yaml.safe_load(f)

if args.output:
    with open(args.output, 'w') as f:
        f.write(t.substitute(y))
else:
    print(t.substitute(y))
