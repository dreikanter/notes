require "pathname"

module PathHelpers
  def file_fixture(fixture_name)
    file_fixture_path.join(fixture_name)
  end

  def file_fixture_path
    root_path.join("spec/fixtures/files")
  end

  def root_path
    Pathname.new(File.expand_path("."))
  end
end

RSpec.configure do |config|
  config.include PathHelpers
end
