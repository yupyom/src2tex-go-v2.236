/* {\hrulefill} *

{\bf Simpson}の公式

{\ % beginning of TeX mode

\noindent
定積分 $\displaystyle\int_a^b f(x)\,dx$ を計算する一つの方法として、
次の近似式が良く使われる。

$$\int_a^b f(x)\,dx\sim{h\over3}\,(y_0+4y_1+2y_2+4y_3+2y_4+4y_5+
          \cdots +2y_{n-2}+4y_{n-1}+y_n)\ ,$$

\noindent
ここで、自然数 $n$ は偶数とし、
$\displaystyle x_i=a+{i\over n}\,(b-a),\ y_i=f(x_i)\quad (i=0, 1, 2, \cdots, n)$
とする。
この公式の証明は、比較的やさしい。
じっさい、Taylor の定理から導かれる等式

$$\int_{\xi-{1\over h}}^{\xi+{1\over h}}f(x)\,dx
  \sim\int_{\xi-{1\over h}}^{\xi+{1\over h}}p(x)\,dx
  ={h\over3}\,\bigl(f(\xi-{1\over h})+4f(\xi)+f(\xi+{1\over h})\bigr)+O(h^5)$$

\noindent
を $\,\displaystyle{n\over2}\,$ 個足し合わせることにより、
この公式は証明される。

{\par\begin{center}\includegraphics[scale=.7]{simpson}\end{center}}

\noindent
ここで、$p(x)$ は次のような条件を満足する２次の多項式を表す:
$$
p\Bigl(\xi-{1\over h}\Bigr)=f\Bigl(\xi-{1\over h}\Bigr),
\ p(\xi)=f(\xi),
\ p\Bigl(\xi+{1\over h}\Bigr)=f\Bigl(\xi+{1\over h}\Bigr)\ .
$$

\noindent
{\tt A, B, F(X), N} の定義をいろいろと変えて、各人で数値実験を行ってみよう。

% end of TeX mode }

* {\hrulefill} */


/* {\hrulefill\ simpson.c\ \hrulefill} */


#include <stdio.h>
#define A 0.                      /* 積分をする区間 $[a, b]$ {\hfill} */
#define B 1.
#define F(X) ((X)*(X))            /* 被積分関数 $f(x)$ {\hfill} */
#define N 20                      /* {\ 区間 $[a, b]$ の分割数 $n$ \hfill} */

double simpson_rule(void)         /* {\ Simpson の公式を用いて $\displaystyle\int_a^b f(x)\,dx$ を計算する関数 \hfill} */
{
    long i;                       /* long 整数を使う {\hfill} */
    double h, s, y[N + 1];        /* {\ メッシュ $h$, 積分値 $s$, 分点 $\displaystyle{i\over n}$ での関数値 $y_i$ \hfill} */

    h = 1. / (double) N;          /* {\ $\displaystyle h={1\over n}$ \hfill} */
    for (i = 0; i <= N; ++i)      /* {\ $\displaystyle y_i=f\bigl({i\over n}\bigr)\quad (i=0, 1, 2, \cdots, n)$ \hfill} */
  y[i] = F((double) i / (double) N);
          /* 以下は {\hfill} */
          /* $\displaystyle
             s={h\over3}(y_0+4y_1+2y_2+4y_3+
             \cdots +4y_{n-1}+y_n)$ {\hfill} */
          /* の計算 {\hfill} */
    s = y[0];
    for (i = 1; i < N; ++i)
    {
  if (i % 2 == 1)                 /* {\ ただし {\tt i \% 2}= $i\ $を 2 で割った余り \hfill} */
      s += 4. * y[i];
  else
      s += 2. * y[i];
    }
    s += y[N];
    s *= (h / 3.);
    return s;                     /* $s$ の値を返す {\hfill} */
}

int main(void)
{
    double s;

    s = simpson_rule();           /* {\ 上で定義した関数simpson\_rule() を使う \hfill} */
    printf("%.16f\n", s);         /* {\ $s$ を小数点以下16桁表示する\hfill} */
    return 0;
}
