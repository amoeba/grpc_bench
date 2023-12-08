#include <fstream>
#include <iostream>

std::string read_file(std::string path) {
  std::ifstream stream(path, std::ios::binary);
  std::string content{std::istreambuf_iterator<char>(stream),
                      std::istreambuf_iterator<char>()};

  stream.close();

  return content;
}

double mean(double *values, int n) {
  double sum = 0;

  for (int i = 0; i < n; i++) {
    sum += values[i];
  }

  return sum / (double)n;
}
