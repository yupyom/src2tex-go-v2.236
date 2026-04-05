# {\hrulefill }
# {\ % beginning of TeX mode }
# {\ \centerline{\bf Towers of Hanoi (Python)} }
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
# {\hrulefill\ hanoi.py \ \hrulefill}

ARRAY = 8                               # {\ disc の数 \hfill }

disc = [[0]*ARRAY for _ in range(3)]
                                        # {\ disc に関するデータの置き場所\hfill }

def init_array():                       # {\ disc に関するデータの初期化\hfill }
    for j in range(ARRAY):
        disc[0][j] = ARRAY - j
        disc[1][j] = 0
        disc[2][j] = 0

counter = 0                             # {\ 移動回数カウンタ \hfill }

def print_result():                     # {\ 結果の表示 \hfill }
    global counter
    counter += 1
    print(f"---#{counter}---")
    for i in range(3):
        print(f"[{i}]", end=" ")
        for j in range(ARRAY):
            if disc[i][j] != 0:
                print(disc[i][j], end=" ")
            else:
                break
        print()

ptr = [0, 0, 0]                         # {\ disc 移動用ポインタ（インデックス）\hfill }

def move_one_disc(i, j):                # {\ 1枚の disc を pole $i$ から\hfill }
                                        # {\ pole $j$ に移動する \hfill }
    ptr[i] -= 1
    disc[j][ptr[j]] = disc[i][ptr[i]]
    ptr[j] += 1
    disc[i][ptr[i]] = 0

def move_discs(n, i, j, k):             # {\ \underline{\textsf{上から $n$ 枚目までの disc}}を、\hfill }
                                        # {\ pole $i$ から pole $j$ に\hfill }
                                        # {\ pole $k$ を経由して、移動する\hfill }
    if n >= 1:
        move_discs(n-1, i, k, j)        # {\ 関数 {\tt move\_discs()}の中で、さらに自分自身 \hfill }
        move_one_disc(i, j)             # {\ {\tt move\_discs()} が使われている。このような \hfill }
        print_result()                  # {\ 手法は、「再帰的呼びだし」といわれる。 \hfill }
        move_discs(n-1, k, j, i)

# {\ \par\begin{center}\includegraphics[scale=0.3]{hanoi1}\quad\includegraphics[scale=0.3]{hanoi2}\end{center} }
#
# {\ たとえば、関数 {\tt move\_discs(4, 0, 1, 2)} を呼び出すことは、 }
# {\ 上図のような操作をすることに対応する。\hfill }

if __name__ == "__main__":
    ptr[0] = ARRAY
    ptr[1] = 0
    ptr[2] = 0

    init_array()
    move_discs(ARRAY, 0, 1, 2)          # {\ {\tt ARRAY} 枚の disc をpole 0 から pole 1 に pole 2\hfill }
                                        # {\ を経由して、移動する \hfill }
