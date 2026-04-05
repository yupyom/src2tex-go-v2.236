; {\hrulefill}
; {\ % beginning of TeX mode}
; {\ \centerline{\bf Towers of Hanoi (Scheme)}}
; {\ \begin{quote}}
; {\ This program gives an answer to the following famous problem (towers of}
; {\ Hanoi).}
; {\ There is a legend that when one of the temples in Hanoi was constructed,}
; {\ three poles were erected and a tower consisting of 64 golden discs was}
; {\ arranged on one pole, their sizes decreasing regularly from bottom to top.}
; {\ The monks were to move the tower of discs to the opposite pole, moving}
; {\ only one at a time, and never putting any size disc above a smaller one.}
; {\ The job was to be done in the minimum numbers of moves. What strategy for}
; {\ moving discs will accomplish this optimum transfer?}
; {\ \end{quote}}{\ % end of TeX mode}{\hrulefill}
; {\hrulefill\ hanoi.scm\ \hrulefill}

(define ARRAY 8)                        ; {\ disc の数 \hfill}

(define disc                            ; {\ disc に関するデータの置き場所\hfill}
    (vector (make-vector ARRAY 0)
            (make-vector ARRAY 0)
            (make-vector ARRAY 0)))

(define (init-array)                    ; {\ disc に関するデータの初期化\hfill}
    (do ((j 0 (+ j 1)))
        ((= j ARRAY))
        (vector-set! (vector-ref disc 0) j (- ARRAY j))
        (vector-set! (vector-ref disc 1) j 0)
        (vector-set! (vector-ref disc 2) j 0)))

(define counter 0)                      ; {\ 移動回数カウンタ \hfill}

(define (print-result)                  ; {\ 結果の表示 \hfill}
    (set! counter (+ counter 1))
    (display "---#") (display counter) (display "---") (newline)
    (do ((i 0 (+ i 1)))
        ((= i 3))
        (display "[") (display i) (display "] ")
        (do ((j 0 (+ j 1)))
            ((or (= j ARRAY)
                 (= (vector-ref (vector-ref disc i) j) 0)))
            (display (vector-ref (vector-ref disc i) j))
            (display " "))
        (newline)))

(define ptr (vector 0 0 0))             ; {\ disc 移動用ポインタ（インデックス）\hfill}

(define (move-one-disc i j)             ; {\ 1枚の disc を pole $i$ から\hfill}
                                        ; {\ pole $j$ に移動する \hfill}
    (vector-set! ptr i (- (vector-ref ptr i) 1))
    (vector-set! (vector-ref disc j) (vector-ref ptr j)
        (vector-ref (vector-ref disc i) (vector-ref ptr i)))
    (vector-set! ptr j (+ (vector-ref ptr j) 1))
    (vector-set! (vector-ref disc i) (vector-ref ptr i) 0))

(define (move-discs n i j k)            ; {\ \underline{\textsf{上から $n$ 枚目までの disc}}を、\hfill}
                                        ; {\ pole $i$ から pole $j$ に\hfill}
                                        ; {\ pole $k$ を経由して、移動する\hfill}
    (cond ((>= n 1)
        (move-discs (- n 1) i k j)      ; {\ 関数 {\tt move-discs}の中で、さらに自分自身 \hfill}
        (move-one-disc i j)             ; {\ {\tt move-discs} が使われている。このような \hfill}
        (print-result)                  ; {\ 手法は、「再帰的呼びだし」といわれる。 \hfill}
        (move-discs (- n 1) k j i))))

; {\ \par\begin{center}\includegraphics[scale=0.3]{hanoi1}\quad\includegraphics[scale=0.3]{hanoi2}\end{center}}
;
; {\ たとえば、関数 {\tt move-discs} を呼び出すことは、}
; {\ 上図のような操作をすることに対応する。\hfill}

(vector-set! ptr 0 ARRAY)
(vector-set! ptr 1 0)
(vector-set! ptr 2 0)

(init-array)
(move-discs ARRAY 0 1 2)                ; {\ {\tt ARRAY} 枚の disc をpole 0 から pole 1 に pole 2\hfill}
                                        ; {\ を経由して、移動する \hfill}
