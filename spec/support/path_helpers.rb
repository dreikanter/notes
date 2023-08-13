require "pathname"

module PathHelpers
  def file_fixture(fixture_name)
    file_fixture_path.join(fixture_name)
  end

  def file_fixture_path
    Pathname.new(File.expand_path(".")).join("spec/fixtures/files")
  end

  RSpec.configure { |config| config.include self }
end
