/* {\hrulefill} *

{\bf Newton-Raphson 法}

方程式 $f(x)=0$ を解くためには、漸化式

$\displaystyle\qquad
  x_0=a,\quad x_{n+1}=x_n-{f(x_n)\over f'(x_n)}
  \qquad\cdots\cdots\ (\star)$

を、適当な初期値 $a$ に対して解いて、数列 $\{x_n\}$ を構成する。
初期値 $a$ がうまく与えられれば、この数列は上述の方程式の
１つの解 $\alpha$ に、収束することが知られている, {\it i.e.},

$\displaystyle\qquad
  \exists\ \alpha=\lim_{n\to\infty}x_n
  \quad such\ that\quad f(\alpha)=0\ .$

漸化式 $(\star)$ の意味とその収束の様子は、次の図と式を見れば
一目瞭然である。

{\par\begin{center}\includegraphics[scale=.7]{newton}\end{center}}

$\displaystyle\qquad
  y-f(x_n)=f'(x_n)(x-x_n)
  \qquad\cdots\cdots\ \hbox{点}\ (x_n,f(x_n))\ \hbox{を通る接線の方程式}$

$\displaystyle\qquad
  x_n-{f(x_n)\over f'(x_n)}
  \qquad\cdots\cdots\ \hbox{上記の接線と}\ x\ \hbox{軸との交点の}
  \ x\ \hbox{座標}$

数列 $\{x_n\}$ の収束性の証明および収束のオーダーの評価は、
それほど簡単なことではない。$\epsilon$-$\delta$ 論法と解析学に関する
知識が必要とされる。興味のある学生には、参考文献を紹介する。

以下のCソース newton.c は、方程式

$\qquad\displaystyle x^2-5=0$

の解の１つが $\sqrt{5}$ であることに注目して、Newton-Raphson 法で $\sqrt{5}$
を求めるプログラムである。{\tt A, F(X), DF(X)} の定義をいろいろ
と変えて、各人で数値実験を行ってみよう。

* {\hrulefill} */


/* {\hrulefill\ newton.c\ \hrulefill} */


#include <stdio.h>
#define A 4.                  /* 初期値 {\hfill} */
#define F(X) ((X)*(X)-5.)     /* 与えられた関数 {\hfill} */
#define DF(X) (2.*(X))        /* その導関数 {\hfill} */

int main(void)
{
    double x, y;              /* 倍精度で計算する {\hfill} */

    x = A;                    /* $\displaystyle x_0=a$ {\hfill} */
    printf("%.16f\n", x);     /* $x_0$ の表示 {\hfill} */
    y = x - F(x) / DF(x);     /* $\displaystyle x_1=x_0 - {f(x_0)\over f'(x_0)}${\hfill} */
    printf("%.16f\n", y);     /* $x_1$ の表示 {\hfill} */
    while (x != y)            /* {\ iteration をやっても値が変わらなくなったら終了\hfill} */
    {
      x = y;
      y = x - F(x) / DF(x);   /* $\displaystyle x_{n+1}=x_n - {f(x_n)\over f'(x_n)}$ {\hfill} */
      printf("%.16f\n", y);   /* $x_{n+1}$ の表示 {\hfill} */
    }
    return 0;
}
