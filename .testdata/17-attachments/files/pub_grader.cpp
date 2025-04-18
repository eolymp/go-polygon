#include <bits/stdc++.h>
using namespace std;

int solve();

int queries = 0;
int n;

int ask(int x) {
    queries++;
    if (n > x) {
        return 1;
    }
    if (n < x) {
        return -1;
    }
    return 0;
}

int main(){
    cin >> n;
    int ans = solve();
    if (queries > 100) {
        cout << "Too many queries" << endl;
    } else if (ans != n) {
        cout << "Incorrect answer" << endl;
    } else {
        cout << queries << endl;
    }
}