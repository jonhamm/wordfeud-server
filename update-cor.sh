#!/usr/bin/env bash
export LC_COLLATE=C
curl -o data/corpus_dk.tsv https://ordregister.dk/files/cor1.02.tsv
cut -f5 data/corpus_dk.tsv | sort -u >data/corpus_dk.txt
wc -l data/*.txt
exit
curl -o tmp/ddo.zip https://korpus.dsl.dk/download/ddo-fullform.zip
unzip -o -d tmp tmp/ddo.zip
mv tmp/ddo_fullforms_*.csv data/ddo.csv
cut -f1 data/ddo.csv >data/ddo.txt

curl -o tmp/ods.zip https://korpus.dsl.dk/download/ods-fullform.zip
unzip -o -d tmp tmp/ods.zip
mv tmp/ods_fullforms_*.csv data/ods.csv
cut -f1 data/ods.csv >data/ods.txt

wc -l data/*.txt
