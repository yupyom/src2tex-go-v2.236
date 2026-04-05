#!/usr/bin/perl
# {\hrulefill }
# {\ % beginning of TeX mode }
# {\ \centerline{\bf Towers of Hanoi (Perl)} }
# {\ \begin{quote} }
# {\ This program gives an answer to the following famous problem (towers of }
# {\ Hanoi). }
# {\ There is a legend that when one of the temples in Hanoi was constructed, }
# {\ three poles were erected and a tower consisting of 64 golden discs was }
# {\ arranged on one pole, their sizes decreasing regularly from bottom to top. }
# {\ The monks were to move the tower of discs to the opposite pole, moving }
# {\ only one at a time, and never putting any size disc above a smaller one. }
# {\ The job was to be done in the minimum numbers of moves. What strategy for }
# {\ moving discs will accomplish this optimum transfer? }
# {\ \end{quote} }{\ % end of TeX mode }{\hrulefill }
# {\hrulefill\ hanoi.pl \ \hrulefill}

use strict;
use warnings;

my $ARRAY = 8;                          # {\ disc の数 \hfill }

my @disc;                               # {\ disc に関するデータの置き場所\hfill }
for my $i (0..2) {
    $disc[$i] = [(0) x $ARRAY];
}

sub init_array {                        # {\ disc に関するデータの初期化\hfill }
    for my $j (0..$ARRAY-1) {
        $disc[0][$j] = $ARRAY - $j;
        $disc[1][$j] = 0;
        $disc[2][$j] = 0;
    }
}

my $counter = 0;                        # {\ 移動回数カウンタ \hfill }
my @ptr = (0, 0, 0);                    # {\ disc 移動用ポインタ（インデックス）\hfill }

sub print_result {                      # {\ 結果の表示 \hfill }
    $counter++;
    print "---#${counter}---\n";
    for my $i (0..2) {
        print "[$i] ";
        for my $j (0..$ARRAY-1) {
            if ($disc[$i][$j] != 0) {
                print "$disc[$i][$j] ";
            } else {
                last;
            }
        }
        print "\n";
    }
}

sub move_one_disc {                     # {\ 1枚の disc を pole $i$ から\hfill }
                                        # {\ pole $j$ に移動する \hfill }
    my ($i, $j) = @_;
    $ptr[$i]--;
    $disc[$j][$ptr[$j]] = $disc[$i][$ptr[$i]];
    $ptr[$j]++;
    $disc[$i][$ptr[$i]] = 0;
}

sub move_discs {                        # {\ \underline{\textsf{上から $n$ 枚目までの disc}}を、\hfill }
                                        # {\ pole $i$ から pole $j$ に\hfill }
                                        # {\ pole $k$ を経由して、移動する\hfill }
    my ($n, $i, $j, $k) = @_;
    if ($n >= 1) {
        move_discs($n - 1, $i, $k, $j);
                                        # {\ 関数 {\tt move\_discs()}の中で、さらに自分自身 \hfill }
        move_one_disc($i, $j);          # {\ {\tt move\_discs()} が使われている。このような \hfill }
        print_result();                 # {\ 手法は、「再帰的呼びだし」といわれる。 \hfill }
        move_discs($n - 1, $k, $j, $i);
    }
}

# {\ \par\begin{center}\includegraphics[scale=0.3]{hanoi1}\quad\includegraphics[scale=0.3]{hanoi2}\end{center} }
#
# {\ たとえば、関数 {\tt move\_discs(4, 0, 1, 2)} を呼び出すことは、 }
# {\ 上図のような操作をすることに対応する。\hfill }

init_array();
$ptr[0] = $ARRAY;
$ptr[1] = 0;
$ptr[2] = 0;

move_discs($ARRAY, 0, 1, 2);            # {\ {\tt ARRAY} 枚の disc をpole 0 から pole 1 に pole 2\hfill }
                                        # {\ を経由して、移動する \hfill }
