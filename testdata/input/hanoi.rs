/* {\hrulefill} *

{\ % beginning of TeX mode

\centerline{\bf Towers of Hanoi (Rust)}

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

/* {\hrulefill\ hanoi.rs \ \hrulefill} */

const ARRAY: usize = 8;                 /* {\ disc の数 \hfill} */

fn init_array(disc: &mut [[i32; ARRAY]; 3]) {
                                        /* {\ disc に関するデータの初期化\hfill} */
    for j in 0..ARRAY {
        disc[0][j] = (ARRAY - j) as i32;
        disc[1][j] = 0;
        disc[2][j] = 0;
    }
}

fn print_result(disc: &[[i32; ARRAY]; 3], counter: &mut i32) {
                                        /* {\ 結果の表示 \hfill} */
    *counter += 1;
    println!("---#{}---", counter);
    for i in 0..3 {
        print!("[{}] ", i);
        for j in 0..ARRAY {
            if disc[i][j] != 0 {
                print!("{} ", disc[i][j]);
            } else {
                break;
            }
        }
        println!();
    }
}

fn move_one_disc(disc: &mut [[i32; ARRAY]; 3], ptr: &mut [usize; 3],
    i: usize, j: usize) {               /* {\ 1枚の disc を pole $i$ からpole $j$ に移動する \hfill} */
    ptr[i] -= 1;
    disc[j][ptr[j]] = disc[i][ptr[i]];
    ptr[j] += 1;
    disc[i][ptr[i]] = 0;
}

fn move_discs(disc: &mut [[i32; ARRAY]; 3], ptr: &mut [usize; 3],
    counter: &mut i32, n: i32, i: usize, j: usize, k: usize) {
                                        /* {\ \underline{\textsf{上から $n$ 枚目までの disc}}を、pole $i$ から pole $j$ に\hfill} */
                                        /* {\ pole $k$ を経由して、移動する\hfill} */
    if n >= 1 {
        move_discs(disc, ptr, counter, n - 1, i, k, j);
                                        /* {\ 関数 {\tt move\_discs()}の中で、さらに自分自身 \hfill} */
        move_one_disc(disc, ptr, i, j); /* {\ {\tt move\_discs()} が使われている。このような \hfill} */
        print_result(disc, counter);    /* {\ 手法は、「再帰的呼びだし」といわれる。 \hfill} */
        move_discs(disc, ptr, counter, n - 1, k, j, i);
    }
}

/* {\par\begin{center}

\includegraphics[scale=0.3]{hanoi1}\quad
\includegraphics[scale=0.3]{hanoi2}\end{center}

たとえば、関数 {\tt move\_discs(4, 0, 1, 2)} を呼び出すことは、
上図のような操作をすることに対応する。\hfill} */

fn main() {
    let mut disc = [[0i32; ARRAY]; 3];  /* {\ disc に関するデータの置き場所\hfill} */
    let mut ptr: [usize; 3] = [ARRAY, 0, 0];
                                        /* {\ disc 移動用ポインタ（インデックス）\hfill} */
    let mut counter: i32 = 0;           /* {\ 移動回数カウンタ \hfill} */

    init_array(&mut disc);
    move_discs(&mut disc, &mut ptr, &mut counter, ARRAY as i32, 0, 1, 2);
                                        /* {\ {\tt ARRAY} 枚の disc をpole 0 から pole 1 に pole 2\hfill} */
                                        /* {\ を経由して、移動する \hfill} */
}
