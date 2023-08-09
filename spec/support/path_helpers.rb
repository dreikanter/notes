require "pathname"

module PathHelpers
  def file_fixture(fixture_name)
    Pathname.new(File.join(file_fixture_path, fixture_name))
  end

  def file_fixture_path
    File.join(root_path, "spec/fixtures/files")
  end

  def root_path
    File.expand_path(".")
  end
end

RSpec.configure do |config|
  config.include PathHelpers
end
