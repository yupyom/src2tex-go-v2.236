#!/bin/bash
# {\hrulefill }
# {\ % beginning of TeX mode }
# {\ \centerline{\bf Towers of Hanoi (Shell)} }
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
# {\hrulefill\ hanoi.sh \ \hrulefill}

ARRAY=8                                 # {\ disc の数 \hfill }
counter=0                               # {\ 移動回数カウンタ \hfill }

init_array() {                          # {\ disc に関するデータの初期化\hfill }
    local j
    for j in $(seq 0 $((ARRAY - 1))); do
        eval "disc_0_$j=$((ARRAY - j))"
        eval "disc_1_$j=0"
        eval "disc_2_$j=0"
    done
    ptr_0=$ARRAY                        # {\ disc 移動用ポインタ（インデックス）\hfill }
    ptr_1=0
    ptr_2=0
}

print_result() {                        # {\ 結果の表示 \hfill }
    local i j val
    counter=$((counter + 1))
    echo "---#${counter}---"
    for i in 0 1 2; do
        printf "[%d] " "$i"
        for j in $(seq 0 $((ARRAY - 1))); do
            eval "val=\$disc_${i}_${j}"
            if [ "$val" -ne 0 ]; then
                printf "%d " "$val"
            else
                break
            fi
        done
        echo
    done
}

move_one_disc() {                       # {\ 1枚の disc を pole $i$ から\hfill }
                                        # {\ pole $j$ に移動する \hfill }
    local i=$1 j=$2 ptr_i ptr_j
    eval "ptr_i=\$ptr_$i"
    ptr_i=$((ptr_i - 1))
    eval "ptr_$i=$ptr_i"
    eval "ptr_j=\$ptr_$j"
    eval "disc_${j}_${ptr_j}=\$disc_${i}_${ptr_i}"
    ptr_j=$((ptr_j + 1))
    eval "ptr_$j=$ptr_j"
    eval "disc_${i}_${ptr_i}=0"
}

move_discs() {                          # {\ \underline{\textsf{上から $n$ 枚目までの disc}}を、\hfill }
                                        # {\ pole $i$ から pole $j$ に\hfill }
                                        # {\ pole $k$ を経由して、移動する\hfill }
    local n=$1 i=$2 j=$3 k=$4
    if [ "$n" -ge 1 ]; then
        move_discs $((n - 1)) "$i" "$k" "$j"
                                        # {\ 関数 {\tt move\_discs()}の中で、さらに自分自身 \hfill }
        move_one_disc "$i" "$j"         # {\ {\tt move\_discs()} が使われている。このような \hfill }
        print_result                    # {\ 手法は、「再帰的呼びだし」といわれる。 \hfill }
        move_discs $((n - 1)) "$k" "$j" "$i"
    fi
}

# {\ \par\begin{center}\includegraphics[scale=0.3]{hanoi1}\quad\includegraphics[scale=0.3]{hanoi2}\end{center} }
#
# {\ たとえば、関数 {\tt move\_discs(4, 0, 1, 2)} を呼び出すことは、 }
# {\ 上図のような操作をすることに対応する。\hfill }

init_array
move_discs $ARRAY 0 1 2                 # {\ {\tt ARRAY} 枚の disc をpole 0 から pole 1 に pole 2\hfill }
                                        # {\ を経由して、移動する \hfill }
