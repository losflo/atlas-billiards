import csv
import sys
import os
import regex

def isfloat(num):
    try:
        val = float(num)
        if val.is_integer():
            val = int(val)
        return val, True
    except ValueError:
        return -1, False

for i in range(1, len(sys.argv)):
    filename = sys.argv[i]
    f = open(filename, 'r')
    fnew = open("formatted/{}".format(filename), 'w')
    r = csv.reader(f, delimiter=',')

    w = csv.writer(fnew, quoting=csv.QUOTE_NONNUMERIC, quotechar='"')
    first = True
    for row in r:
        if first:
            first = False
            w.writerow(row)
            continue

        for i in range(len(row)):
            num, ok = isfloat(row[i])
            if ok:
                row[i] = num
        w.writerow(row)
    f.close()

