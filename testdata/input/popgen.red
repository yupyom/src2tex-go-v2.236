% {\bf popgen.red}
% {\null
% We shall solve the equation
% $$
% \eqalign{
% {\partial u\over\partial t}
%	&=a(x){\partial^2 u\over\partial x^2}
%	+b(x){\partial u\over\partial x}
%	+c(x) u\cr
%	&={x(1-x)\over 2}{\partial^2 u\over\partial x^2}
%	+b(x){\partial u\over\partial x}
%	+c(x) u\cr
% }
% $$
% by using numerical-symbolic hybrid method established by K. Amano.
% This equation plays an important role in population genetics.
% }

% coefficients
procedure sqrt_a(x);
   sqrt(x*(1-x)/2);
procedure b(x);
   1/2-x;
procedure c(x);
   0;

% domain
% {\null
% $$
% {\tt domain\_p(t,x)}=\cases{
%	1 &if $(t,x)$ belongs to the domain\cr
%	\cr
%	0 &otherwise\cr
% }
% $$
% }
procedure domain_p(t,x);
begin;
   on rounded;
   if t < 0 or x < 0 or x > 1 then
   <<
      off rounded;
      return 0
   >>
   else
   <<
      off rounded;
      return 1
   >>
end;

% numerical-symbolic hybrid method
% {\ Key idea depends on the following formula:
% $$
% \eqalign{
% u(t,x)
% &={1\over6}\,u\bigl(t,x+\sqrt{a(x)}\,h\bigr)
%	+{1\over6}\,u\bigl(t,x-\sqrt{a(x)}\,h\bigr)\cr
% &+{1\over3}\,u\bigl(t,x+b(x)h^2\bigr)
%	+{1\over3}\,u(t-h^2,x)
%	+{h^2\over3}\,c(x)u(t,x)
%	+O(h^4)\ .\cr
% }
% $$
% }
procedure hybrid_method(t, x, n);
begin;
   list_in := {{1, t, x}};
   list_tmp := {};
   while n > 0 do
   <<
      while length(list_in) > 0 do
      <<
         tmp := first(list_in);
         q := first(tmp);
         s := first(rest(tmp));
         y := first(rest(rest(tmp)));
         if domain_p(s, y) neq 0 then
         <<
            list_tmp := cons({q/6, s, y+sqrt_a(y)*h}, list_tmp);
            list_tmp := cons({q/6, s, y-sqrt_a(y)*h}, list_tmp);
            list_tmp := cons({q/3, s, y+b(y)*h**2}, list_tmp);
            list_tmp := cons({q/3, s-h**2, y}, list_tmp);
            list_tmp := cons({h**2*c(y)/3, s, y}, list_tmp)
         >>
         else
            list_tmp := cons({q, s, y}, list_tmp);
         list_in := rest(list_in)
      >>;
      list_in := list_tmp;
      list_tmp := {};
      n := n-1
   >>;
   return list_in
end;

% main module
h := 0.1;
hybrid_method(4, 0.5, 2);

end;

% Here we give a numerical example.
% {\null
% $$
% \cases{
% \displaystyle{\partial u\over\partial t}={1\over4N}
% {\partial^2\over\partial x^2}\big(x(1-x)u\big)
%	\qquad (0<t\le 4N,\ 0<x<1)\cr
% \cr
%u(0,x)\sim\delta(x-0.5)\cr
% }
% $$
% }

% {\special{epsfile=solution.eps hscale=0.7 vscale=0.7}}

