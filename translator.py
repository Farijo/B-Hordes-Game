from string import Template
from os.path import basename, join
import argparse
import yaml

parser = argparse.ArgumentParser(
                    prog=basename(__file__),
                    description='Generate a translated file based on template_file for each TRANSLATION_FILE')
parser.add_argument('template_file')
parser.add_argument('-tr', '--translation-file', nargs="+", help='Yaml translation files to apply')
parser.add_argument('-trname', '--translate-name', action='store_true', help='Also apply translation to filename')
parser.add_argument('-o', '--output-dir', help='Directory to write generated files', default='.')
parser.add_argument('-d', '--delimiter', help='Custom delimiter for template parsing', default='@@')
args = parser.parse_args()

class GoTemplate(Template):
    delimiter = args.delimiter
    idpattern = '(?a:[\-_a-z][\-_a-z0-9]*)'

outfile = basename(args.template_file)

with open(args.template_file) as f:
    t = GoTemplate(f.read())

for lng in args.translation_file:
    with open(lng) as f:
        y = yaml.safe_load(f)

    genfilename = outfile
    if args.translate_name:
        genfilename = GoTemplate(genfilename).substitute(y)

    with open(join(args.output_dir, genfilename), 'w') as f:
        f.write(t.substitute(y))
