require 'rake'
require 'date'
require 'rubygems/package'
require 'term/ansicolor'
color = Term::ANSIColor
app_name = "goaway"
release = DateTime.now.strftime('%Y-%m-%d-%H%M%S')
tarfile_name = "#{app_name}-#{release}.tar.gz"

task :build do
  sh 'glide up'
  sh 'go build'
  sh "./goaway build-config #{ENV['ENV']}"
  sh 'mkdir -p build'
  sh 'mv goaway build/goaway'
  sh 'mv conf/goaway.conf build/goaway.conf'
end

task :package => 'build' do
  sh "tar -cf #{tarfile_name} build"
  sh 'rm -rf build'
  sh "mv #{tarfile_name} ../../../"
end

desc 'new build'
multitask :all => ['build', 'package']
