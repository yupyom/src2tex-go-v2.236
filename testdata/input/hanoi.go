/* {\hrulefill} *

{\ % beginning of TeX mode

\centerline{\bf Towers of Hanoi (Go)}

\begin{quote}

This program gives an answer to the following famous problem (towers of
Hanoi).
There is a legend that when one of the temples in Hanoi was constructed,
three poles were erected and a tower consisting of 64 golden discs was
arranged on one pole, their sizes decreasing regularly from bottom to top.
The monks were to move the tower of discs to the opposite pole, moving
only one at a time, and never putting any size disc above a smaller one.
The job was to be done in the minimum numbers of moves. What strategy for
moving discs will accomplish this optimum transfer?
\end{quote}

% end of TeX mode }

* {\hrulefill} */

/* {\hrulefill\ hanoi.go \ \hrulefill} */

package main

import "fmt"

const ARRAY = 8                         /* {\ disc の数 \hfill} */

var disc [3][ARRAY]int                  /* {\ disc に関するデータの置き場所\hfill} */

func initArray() {                      /* {\ disc に関するデータの初期化\hfill} */
    for j := 0; j < ARRAY; j++ {
        disc[0][j] = ARRAY - j
        disc[1][j] = 0
        disc[2][j] = 0
    }
}

var counter int                         /* {\ 移動回数カウンタ \hfill} */

func printResult() {                    /* {\ 結果の表示 \hfill} */
    counter++
    fmt.Printf("---#%d---\n", counter)
    for i := 0; i <= 2; i++ {
        fmt.Printf("[%d] ", i)
        for j := 0; j < ARRAY; j++ {
            if disc[i][j] != 0 {
                fmt.Printf("%d ", disc[i][j])
            } else {
                break
            }
        }
        fmt.Println()
    }
}

var ptr [3]int                          /* {\ disc 移動用ポインタ（インデックス）\hfill} */

func moveOneDisc(i, j int) {            /* {\ 1枚の disc を pole $i$ からpole $j$ に移動する \hfill} */
    ptr[i]--
    disc[j][ptr[j]] = disc[i][ptr[i]]
    ptr[j]++
    disc[i][ptr[i]] = 0
}

func moveDiscs(n, i, j, k int) {        /* {\ \underline{\textsf{上から $n$ 枚目までの disc}}を、pole $i$ から pole $j$ に\hfill} */
                                        /* {\ pole $k$ を経由して、移動する\hfill} */
    if n >= 1 {
        moveDiscs(n-1, i, k, j)         /* {\ 関数 {\tt moveDiscs()}の中で、さらに自分自身 \hfill} */
        moveOneDisc(i, j)               /* {\ {\tt moveDiscs()} が使われている。このような \hfill} */
        printResult()                   /* {\ 手法は、「再帰的呼びだし」といわれる。 \hfill} */
        moveDiscs(n-1, k, j, i)
    }
}

/* {\par\begin{center}

\includegraphics[scale=0.3]{hanoi1}\quad
\includegraphics[scale=0.3]{hanoi2}\end{center}

たとえば、関数 {\tt moveDiscs(4, 0, 1, 2)} を呼び出すことは、
上図のような操作をすることに対応する。\hfill} */

func main() {
    ptr[0] = ARRAY
    ptr[1] = 0
    ptr[2] = 0

    initArray()
    moveDiscs(ARRAY, 0, 1, 2)           /* {\ {\tt ARRAY} 枚の disc をpole 0 から pole 1 に pole 2\hfill} */
                                        /* {\ を経由して、移動する \hfill} */
}
