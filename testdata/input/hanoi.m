% {\hrulefill}
% {\ % beginning of TeX mode}
% {\ \centerline{\bf Towers of Hanoi (Octave)}}
% {\ \begin{quote}}
% {\ This program gives an answer to the following famous problem (towers of}
% {\ Hanoi).}
% {\ There is a legend that when one of the temples in Hanoi was constructed,}
% {\ three poles were erected and a tower consisting of 64 golden discs was}
% {\ arranged on one pole, their sizes decreasing regularly from bottom to top.}
% {\ The monks were to move the tower of discs to the opposite pole, moving}
% {\ only one at a time, and never putting any size disc above a smaller one.}
% {\ The job was to be done in the minimum numbers of moves. What strategy for}
% {\ moving discs will accomplish this optimum transfer?}
% {\ \end{quote}}{\ % end of TeX mode}{\hrulefill}
% {\hrulefill\ hanoi.m \ \hrulefill}

global ARRAY;
ARRAY = 8;                              % {\ disc の数 \hfill}

global disc;
disc = zeros(3, ARRAY);                 % {\ disc に関するデータの置き場所\hfill}

function init_array()                   % {\ disc に関するデータの初期化\hfill}
    global ARRAY disc;
    for j = 1:ARRAY
        disc(1, j) = ARRAY - j + 1;
        disc(2, j) = 0;
        disc(3, j) = 0;
    endfor
endfunction

global counter;
counter = 0;                            % {\ 移動回数カウンタ \hfill}

function print_result()                 % {\ 結果の表示 \hfill}
    global ARRAY disc counter;          % {\ Octave は 1-indexed のため\hfill}
    counter = counter + 1;              % {\ ptr の値は実際の位置 $+1$ で管理する\hfill}
    printf("---#%d---\n", counter);
    for i = 1:3
        printf("[%d] ", i - 1);
        for j = 1:ARRAY
            if disc(i, j) != 0
                printf("%d ", disc(i, j));
            else
                break;
            endif
        endfor
        printf("\n");
    endfor
endfunction

global ptr;
ptr = ones(1, 3);                       % {\ disc 移動用ポインタ（1-indexed）\hfill}

function move_one_disc(i, j)            % {\ 1枚の disc を pole $i$ から\hfill}
                                        % {\ pole $j$ に移動する \hfill}
    global disc ptr;
    ptr(i) = ptr(i) - 1;
    disc(j, ptr(j)) = disc(i, ptr(i));
    ptr(j) = ptr(j) + 1;
    disc(i, ptr(i)) = 0;
endfunction

function move_discs(n, i, j, k)         % {\ \underline{\textsf{上から $n$ 枚目までの disc}}を、\hfill}
                                        % {\ pole $i$ から pole $j$ に\hfill}
                                        % {\ pole $k$ を経由して、移動する\hfill}
    if n >= 1
        move_discs(n - 1, i, k, j);     % {\ 関数 {\tt move\_discs()}の中で、さらに自分自身 \hfill}
        move_one_disc(i, j);            % {\ {\tt move\_discs()} が使われている。このような \hfill}
        print_result();                 % {\ 手法は、「再帰的呼びだし」といわれる。 \hfill}
        move_discs(n - 1, k, j, i);
    endif
endfunction

% {\par\begin{center}\includegraphics[scale=0.3]{hanoi1}\quad\includegraphics[scale=0.3]{hanoi2}\end{center}}
%
% {\ たとえば、関数 {\tt move\_discs(4, 1, 2, 3)} を呼び出すことは、}
% {\ 上図のような操作をすることに対応する。\hfill}

ptr(1) = ARRAY + 1;                     % {\ 1-indexed のため ARRAY $+1$ \hfill}
ptr(2) = 1;
ptr(3) = 1;

init_array();
move_discs(ARRAY, 1, 2, 3);             % {\ {\tt ARRAY} 枚の disc をpole 0 から pole 1 に pole 2\hfill}
                                        % {\ を経由して、移動する \hfill}
