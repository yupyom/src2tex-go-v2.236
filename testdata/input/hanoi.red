% {\hrulefill}
% {\ % beginning of TeX mode}
% {\ \centerline{\bf Towers of Hanoi (Reduce)}}
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
% {\hrulefill\ hanoi.red \ \hrulefill}

array_size := 8;                        % {\ disc の数 \hfill}

array disc(2, 7);                       % {\ disc に関するデータの置き場所\hfill}

procedure init_array();                 % {\ disc に関するデータの初期化\hfill}
begin;
    for j := 0:array_size-1 do
    <<
        disc(0, j) := array_size - j;
        disc(1, j) := 0;
        disc(2, j) := 0
    >>
end;

counter := 0;                           % {\ 移動回数カウンタ \hfill}

procedure print_result();               % {\ 結果の表示 \hfill}
begin;
    counter := counter + 1;
    write "---#", counter, "---";
    terpri();
    for i := 0:2 do
    <<
        write "[", i, "] ";
        for j := 0:array_size-1 do
            if disc(i, j) neq 0 then
                write disc(i, j), " "
            else
                j := array_size;        % {\ break 相当 \hfill}
        terpri()
    >>
end;

array ptr(2);                           % {\ disc 移動用ポインタ（インデックス）\hfill}

procedure move_one_disc(i, j);          % {\ 1枚の disc を pole $i$ から\hfill}
                                        % {\ pole $j$ に移動する \hfill}
begin;
    ptr(i) := ptr(i) - 1;
    disc(j, ptr(j)) := disc(i, ptr(i));
    ptr(j) := ptr(j) + 1;
    disc(i, ptr(i)) := 0
end;

procedure move_discs(n, i, j, k);       % {\ \underline{\textsf{上から $n$ 枚目までの disc}}を、\hfill}
                                        % {\ pole $i$ から pole $j$ に\hfill}
                                        % {\ pole $k$ を経由して、移動する\hfill}
begin;
    if n >= 1 then
    <<
        move_discs(n - 1, i, k, j);     % {\ 関数 {\tt move\_discs()}の中で、さらに自分自身 \hfill}
        move_one_disc(i, j);            % {\ {\tt move\_discs()} が使われている。このような \hfill}
        print_result();                 % {\ 手法は、「再帰的呼びだし」といわれる。 \hfill}
        move_discs(n - 1, k, j, i)
    >>
end;

% {\par\begin{center}\includegraphics[scale=0.3]{hanoi1}\quad\includegraphics[scale=0.3]{hanoi2}\end{center}}
%
% {\ たとえば、関数 {\tt move\_discs(4, 0, 1, 2)} を呼び出すことは、}
% {\ 上図のような操作をすることに対応する。\hfill}

ptr(0) := array_size;
ptr(1) := 0;
ptr(2) := 0;

init_array();
move_discs(array_size, 0, 1, 2);        % {\ {\tt array\_size} 枚の disc をpole 0 から pole 1 に pole 2\hfill}
                                        % {\ を経由して、移動する \hfill}

end;
