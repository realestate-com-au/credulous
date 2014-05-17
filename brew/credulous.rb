require 'formula'

class Credulous < Formula
  homepage 'https://github.com/realestate-com-au/credulous'
  version '0.1.2'

  url "https://github.com/realestate-com-au/credulous/releases/download/#{version}/credulous-#{version}-osx"
  sha1 '92b7485cee761ff1a4fff2f20b46a2311a2d7302'
  
  def install
    bin.install "credulous-#{version}-osx"
    system "mv "+bin+"/credulous-#{version}-osx "+bin+"/credulous"
  end

  test do
    assert_equal "Credulous version #{version}", `#{bin}/credulous -v`.strip
  end

end
